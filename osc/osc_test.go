package osc

import (
	"bytes"
	// "fmt"
	"testing"
	"time"
)

func TestAddArgs(t *testing.T) {

	elements := []interface{}{
		int32(123),
		"tester",
		float32(4.5),
		nil,
		true,
		false,
	}

	m := &Message{Address: "/add/args"}

	for _, e := range elements {
		m.Args = append(m.Args, e)
	}

	if len(m.Args) != len(elements) {
		t.Error("arg length is incorrect!")
	}

	for i, a := range m.Args {
		if a != elements[i] {
			t.Errorf("added arg doesn't match, %v != %v", a, elements[i])
		}
	}
}

func TestWriteArgs(t *testing.T) {

	m := &Message{Address: "/write/args"}
	m.Args = append(m.Args, "test1")
	m.Args = append(m.Args, int32(345))
	m.Args = append(m.Args, float32(34.5))

	buf := new(bytes.Buffer)
	numbytes, e := m.WriteTo(buf)
	if e != nil {
		t.Error("Error writing to buf: ", e)
	}

	buflen := len(buf.Bytes())

	if numbytes != buflen {
		t.Errorf("incorrect number of bytes reported written. %d reported, actual %d:", numbytes, buflen)
	}

	if (buflen & 3) != 0 {
		t.Error("written buffer size was not 4-byte aligned, len:", buflen)
	}
}

func TestRoundTrip(t *testing.T) {

	m := &Message{Address: "/round/trip"}
	m.Args = append(m.Args, int32(345))
	m.Args = append(m.Args, float32(34.5))
	m.Args = append(m.Args, "monkey")
	m.Args = append(m.Args, true)
	m.Args = append(m.Args, false)
	m.Args = append(m.Args, nil)

	// test a non-4-byte aligned blob size to make sure
	// padding is handled correctly
	blob := []byte{0x1, 0x2, 0x3, 0x4, 0x5}
	m.Args = append(m.Args, blob)

	buf := new(bytes.Buffer)
	numbytes, e := m.WriteTo(buf)

	if e != nil {
		t.Error("Error writing to buf: ", e)
	}

	if numbytes != buf.Len() {
		t.Error("Message.WriteTo() reported incorrect length")
	}

	if (numbytes & 3) != 0 {
		t.Error("Message: written buffer size was not 4-byte aligned, len:", numbytes)
	}

	bndl, err := ReadFrom(buf)
	if err != nil {
		t.Error("Error reading from buf: ", err)
	}

	if len(bndl.Messages) != 1 {
		t.Error("Incorrect number of messages read:", len(bndl.Messages))
	}

	if !Equal(m, bndl.Messages[0]) {
		t.Error("messages are not equal")
	}
}

func TestTimetag(t *testing.T) {

	now := time.Now()
	tt := timeToTimetag(now)
	roundtrip := timetagToTime(tt)

	// we're not guaranteed to retain complete precision for all
	// incoming timetags, but at least for roundtrips within
	// the same environment we should match

	if now != roundtrip {
		t.Error("times don't match:", now, roundtrip)
	}
}

func TestBundleRoundtrip(t *testing.T) {

	now := time.Now()

	b := &Bundle{Timetag: now}

	m0 := &Message{Address: "/bundle/round/trip/0"}
	m0.Args = append(m0.Args, int32(345))
	m0.Args = append(m0.Args, float32(34.5))

	m1 := &Message{Address: "/bundle/round/trip/1"}
	m1.Args = append(m1.Args, int32(567))
	m1.Args = append(m1.Args, float32(56.7))

	b.Messages = append(b.Messages, m0)
	b.Messages = append(b.Messages, m1)

	buf := new(bytes.Buffer)
	numbytes, e := b.WriteTo(buf)

	if e != nil {
		t.Error("Error writing Bundle to buf:", e)
	}

	if numbytes != buf.Len() {
		t.Error("Message.WriteTo() reported incorrect length")
	}

	if (numbytes & 3) != 0 {
		t.Error("Bundle: written buffer size was not 4-byte aligned, len:", numbytes)
	}

	br, err := ReadFrom(buf)

	if err != nil {
		t.Error("Error reading Bundle from buf:", e)
	}

	if br.Timetag != now {
		t.Error("timetags don't match:", br.Timetag, now)
	}

	if len(br.Messages) != 2 {
		t.Error("incorrect number of messages found in Bundle:", len(br.Messages))
	}

	for i, msg := range b.Messages {
		if !Equal(msg, br.Messages[i]) {
			t.Error("messages are not equal")
		}
	}
}
