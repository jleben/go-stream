// gosc provides a pure Go package for sending and receiving OpenSoundControl
// messages.
//
// Sending messages is a matter of writing a Message to the appropriate
// io.Writer (net/UDPConn is a common transport), and a simple UDPServer
// is provided for receiving Messages.
//
// Address pattern matching is also supported.
//
// See the OSC spec at http://opensoundcontrol.org/spec-1_0 for more detail.
//
package osc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	// The time tag value consisting of 63 zero bits followed by a one
	// in the least signifigant bit is a special case meaning "immediately."
	timetagImmediate      = uint64(1)
	secondsFrom1900To1970 = 2208988800
)

// A single OSC message.
// Args is a slice of arguments, each of which may be one of
// the following types: int32, float32, string, bool, nil,
// or []byte (ie, an OSC 'blob')
type Message struct {
	Address string
	Args    []interface{}
}

// A collection of OSC Messages.
// It's often more efficient to send multiple messages as a Bundle
// as they can all be included in the same transmission to the receiver.
type Bundle struct {
	Timetag  time.Time
	Messages []*Message
}

// WriteTo writes b to w.
func (b *Bundle) WriteTo(w io.Writer) (n int, err error) {

	writer := bufio.NewWriter(w)

	n, err = writePaddedString("#bundle", writer)
	if err != nil {
		return 0, err
	}

	// write timetag
	tt := timeToTimetag(b.Timetag)
	if err = binary.Write(writer, binary.BigEndian, &tt); err != nil {
		return 0, err
	}

	n = 16 // 8 bytes of "#bundle" and 8 bytes of timetag

	// each message must be preceded by its int32 length
	for _, m := range b.Messages {

		var ntmp int
		var buf bytes.Buffer

		if ntmp, err = m.WriteTo(&buf); err != nil {
			return 0, err
		}

		numbytes := int32(ntmp)
		if err = binary.Write(writer, binary.BigEndian, &numbytes); err != nil {
			return 0, err
		}

		if ntmp, err = writer.Write(buf.Bytes()); err != nil {
			return 0, err
		}

		n += ntmp + 4
	}

	return n, writer.Flush()
}

// WriteTo writes m to w.
func (m *Message) WriteTo(w io.Writer) (n int, err error) {

	writer := bufio.NewWriter(w)

	// we can write out the address string immediately
	n, err = writePaddedString(m.Address, writer)
	if err != nil {
		return 0, err
	}

	// typetag starts with ,
	typetag := []byte{','}

	// now we can write out the typetag and collect the payload
	// as we parse the args
	var payload bytes.Buffer

	for _, arg := range m.Args {
		switch t := arg.(type) {

		default:
			return 0, errors.New(fmt.Sprintf("osc - unsupported type: %T", t))

		case bool:
			if arg.(bool) == true {
				typetag = append(typetag, 'T')
			} else {
				typetag = append(typetag, 'F')
			}

		case nil:
			typetag = append(typetag, 'N')

		case int32:
			typetag = append(typetag, 'i')

			if err = binary.Write(&payload, binary.BigEndian, arg); err != nil {
				return 0, err
			}

		case float32:
			typetag = append(typetag, 'f')

			if err = binary.Write(&payload, binary.BigEndian, arg); err != nil {
				return 0, err
			}

		case string:
			typetag = append(typetag, 's')

			if _, err = writePaddedString(arg.(string), &payload); err != nil {
				return 0, err
			}

		case []byte:
			typetag = append(typetag, 'b')
			if _, err = writeBlob(arg.([]byte), &payload); err != nil {
				return 0, err
			}

		}
	}

	var ntmp int // holder for incremental write lengths

	if ntmp, err = writePaddedString(string(typetag), writer); err != nil {
		return 0, err
	}
	n += ntmp

	if ntmp, err = writer.Write(payload.Bytes()); err != nil {
		return 0, err
	}
	n += ntmp

	return n, writer.Flush()
}

// ReadFrom retrieves a Bundle from the given io.Reader.
// In the event that the incoming packet is a single message,
// the returned Bundle will contain a single Message with its
// Timetag set to 'immediate'
func ReadFrom(r io.Reader) (bndl *Bundle, err error) {

	reader := bufio.NewReader(r)

	var peekBytes []byte
	if peekBytes, err = reader.Peek(1); err != nil {
		return bndl, err
	}

	if peekBytes[0] == '/' {
		// single messages don't supply a timetag - assume they're immediate
		bndl = &Bundle{Timetag: timetagToTime(timetagImmediate)}

		var m *Message
		if m, err = readOneMessage(reader); err == nil {
			bndl.Messages = append(bndl.Messages, m)
		}
		return bndl, err

	} else if peekBytes[0] == '#' {
		// receiving a bundle
		var bndlstr string
		if bndlstr, err = readPaddedString(reader); err != nil {
			return nil, err
		}

		if bndlstr != "#bundle" {
			return nil, errors.New("incorrect bundle header")
		}

		var tt uint64
		if err = binary.Read(reader, binary.BigEndian, &tt); err != nil {
			return nil, err
		}

		bndl = &Bundle{Timetag: timetagToTime(tt)}

		// messages are back-to-back, preceded by their int32 length
		for {
			var mlen int32
			if err = binary.Read(reader, binary.BigEndian, &mlen); err != nil {
				if err == io.EOF {
					return bndl, nil // successfully reached the end of the bundle
				}
				return nil, err
			}

			// XXX: validate that num bytes read match mlen?

			var m *Message
			if m, err = readOneMessage(reader); err == nil {
				bndl.Messages = append(bndl.Messages, m)
			} else {
				return nil, err
			}
		}
	}

	return bndl, errors.New("packets must start with '/' or '#'")
}

// Equal returns a boolean reporting whether a == b, by comparing
// the Address and the contents of all Args
func Equal(a, b *Message) bool {

	if a.Address != b.Address {
		return false
	}

	if len(a.Args) != len(b.Args) {
		return false
	}

	for i, arg := range a.Args {
		switch arg.(type) {

		case bool, nil, int32, float32, string:
			if arg != b.Args[i] {
				return false
			}
		case []byte:
			ba := arg.([]byte)
			bb := b.Args[i].([]byte)
			if !bytes.Equal(ba, bb) {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// unify bytes.Buffer and bufio.Reader for our purposes
type bufReader interface {
	Read(p []byte) (n int, err error)
	ReadString(delim byte) (line string, err error)
}

func readOneMessage(r bufReader) (msg *Message, err error) {

	msg = new(Message)

	if msg.Address, err = readPaddedString(r); err != nil {
		return msg, err
	}

	if err = readArgs(msg, r); err != nil {
		return msg, err
	}

	return msg, nil
}

func readArgs(msg *Message, r bufReader) (err error) {

	var typetag string
	if typetag, err = readPaddedString(r); err != nil {
		return err
	}

	if typetag[0] != ',' {
		return errors.New("illegal typetag")
	}

	for _, c := range typetag[1:] {
		switch c {
		default:
			return errors.New(fmt.Sprintf("unsupported type in typetag: %s", c))

		case 'T':
			msg.Args = append(msg.Args, true)

		case 'F':
			msg.Args = append(msg.Args, false)

		case 'N':
			msg.Args = append(msg.Args, nil)

		case 'i':
			var i int32
			if err = binary.Read(r, binary.BigEndian, &i); err != nil {
				return err
			}
			msg.Args = append(msg.Args, i)

		case 'f':
			var f float32
			if err = binary.Read(r, binary.BigEndian, &f); err != nil {
				return err
			}
			msg.Args = append(msg.Args, f)

		case 's':
			var s string
			if s, err = readPaddedString(r); err != nil {
				return err
			}
			msg.Args = append(msg.Args, s)

		case 'b':
			var blob []byte
			if blob, err = readBlob(r); err != nil {
				return err
			}
			msg.Args = append(msg.Args, blob)
		}
	}

	return nil
}

func padBytesNeeded(elementLen int) int {
	return 4*(elementLen/4+1) - elementLen
}

// unify bytes.Buffer and bufio.Writer for our purposes
type bufWriter interface {
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
}

func writePaddedString(s string, b bufWriter) (n int, err error) {

	if n, err = b.WriteString(s); err != nil {
		return 0, err
	}

	padlen := padBytesNeeded(len(s))
	// contets of a new array are already 0
	padding := make([]byte, padlen)
	if padlen, err = b.Write(padding); err != nil {
		return 0, err
	}

	return n + padlen, nil
}

func readPaddedString(r bufReader) (s string, err error) {

	if s, err = r.ReadString(0); err != nil {
		return "", err
	}

	// bufio.Reader.ReadString includes the delimiting 0 in the
	// returned string so we need to remove it and account for it
	// when calculating the amount of padding

	s = s[:len(s)-1]

	// slurp out extra padding
	padlen := padBytesNeeded(len(s)) - 1
	if padlen > 0 {
		dummy := make([]byte, padlen)
		if _, err = r.Read(dummy); err != nil {
			return s, err
		}
	}

	return s, nil
}

// Blobs are int32 length, followed by length bytes, followed by pad bytes
// to ensure total size is 32-bit aligned, even if length is not

func writeBlob(blob []byte, b bufWriter) (n int, err error) {

	blen := int32(len(blob))
	if err = binary.Write(b, binary.BigEndian, &blen); err != nil {
		return 0, err
	}

	if _, err = b.Write(blob); err != nil {
		return 0, err
	}

	padlen := padBytesNeeded(int(blen))
	if padlen > 0 {
		padding := make([]byte, padlen)
		if padlen, err = b.Write(padding); err != nil {
			return 0, err
		}
	}

	return 4 + int(blen) + padlen, nil
}

func readBlob(r bufReader) (blob []byte, err error) {

	var blen int32
	if err = binary.Read(r, binary.BigEndian, &blen); err != nil {
		return nil, err
	}

	blob = make([]byte, blen)
	if _, err = r.Read(blob); err != nil {
		return nil, err
	}

	padlen := padBytesNeeded(int(blen))
	if padlen > 0 {
		dummy := make([]byte, padlen)
		if _, err = r.Read(dummy); err != nil {
			return nil, err
		}
	}

	return blob, nil
}

// Time tags are represented by a 64 bit fixed point number.
// The first 32 bits specify the number of seconds since midnight on
// January 1, 1900, and the last 32 bits specify fractional parts of
// a second to a precision of about 200 picoseconds.
// This is therepresentation used by Internet NTP timestamps.
//
// The time tag value consisting of 63 zero bits followed by a one in the
// least signifigant bit is a special case meaning "immediately."

func timeToTimetag(t time.Time) (timetag uint64) {
	timetag = uint64((secondsFrom1900To1970 + t.Unix()) << 32)
	return timetag + uint64(uint32(t.Nanosecond()))
}

func timetagToTime(timetag uint64) (t time.Time) {
	return time.Unix(int64((timetag>>32)-secondsFrom1900To1970), int64(timetag&0xffffffff))
}
