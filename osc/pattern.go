package osc

import (
	// "fmt"
	"strings"
)

// Determines whether pattern (typically the Address of an
// incoming Message) matches test (typically the Address of a registered
// handler).
//
// Pattern matching semantics are described in the OSC spec
// (http://opensoundcontrol.org/spec-1_0) in the section entitled
// "OSC Message Dispatching and Pattern Matching".
//
func PatternMatch(pattern, test string) bool {

	// indexes into our strings
	ipattern := 0
	itest := 0

	for {
		// check for end of string, or just empty string
		if ipattern == len(pattern) {
			return itest == len(test)
		}

		c := pattern[ipattern]

		// check for end/empty test string
		if itest == len(test) {
			// if there's anything else in pattern other than * or escape char, no match
			if c != '*' && c != '\\' {
				return false
			}

			ipattern++
			continue
		}

		switch c {
		case '?':
			// matches any single character
			ipattern++
			itest++

		case '*':
			result, poffset, toffset := matchWildcard(pattern[ipattern:], test[itest:])
			if !result {
				return false
			}
			ipattern += poffset
			itest += toffset

		case ']', '}':
			// OSCWarning("Spurious %c in pattern \".../%s/...\"",pattern[0], theWholePattern);
			return false

		case '[':
			result, numbytes := matchSet(pattern[ipattern:], rune(test[itest]))
			if !result {
				return false
			}
			ipattern += numbytes
			itest++

		case '{':
			// implement me
			return false

		case '\\':
			// special case if this is the last character
			if ipattern == len(pattern)-1 {
				return itest == len(test)
			}

			// if the following character doesn't match, we're done
			if pattern[ipattern+1] != test[itest] {
				return false
			}

			// continue via basic character match after the escaped character
			ipattern += 2
			itest++

		default:
			// basic character match - if these two characters match,
			// continue to the next character, otherwise we're done
			if c != test[itest] {
				return false
			}

			ipattern++
			itest++
		}
	}

	panic("unreachable - failure in osc.PatternMatch()")
}

// '*' in the OSC Address Pattern matches any sequence of zero or more characters
//
// we assume it's not meaningful to have any other 'special' characters in the
// pattern string beyond the wildcard - it has already matched everything up
// to the next / delimited part anyway
func matchWildcard(pattern, test string) (bool, int, int) {

	// check for next part delimiter
	pidx := strings.IndexRune(pattern, '/')
	tidx := strings.IndexRune(test, '/')

	// both have one - move each to it
	if pidx >= 0 && tidx >= 0 {
		return true, pidx, tidx
	}

	// neither have one - move each to their end
	if pidx < 0 && tidx < 0 {
		return true, len(pattern), len(test)
	}

	// mismatched number of parts
	return false, 0, 0
}

// spec: "a string of characters in square brackets (e.g., "[string]")
// in the OSC Address Pattern matches any character in the string."
func matchSet(pattern string, test rune) (bool, int) {

	// "An exclamation point at the beginning of a bracketed string
	// negates the sense of the list, meaning that the list matches
	// any character not in the list. (An exclamation point anywhere
	// besides the first character after the open bracket has no special meaning.)"
	var negated bool
	if pattern[1] == '!' {
		negated = true
		pattern = pattern[1:]
	} else {
		negated = false
	}

	for i, c := range pattern {
		// if we got to the closing bracket without matching,
		// 'negated' specifies whether that's what was actually asked for
		if c == ']' {
			// account for one byte of '!' and one byte to step past the current ']'
			return negated, i + 2
		}

		// "two characters separated by a minus sign indicate the
		// range of characters between the given two in ASCII collating sequence."
		// ie, check for 'a-z' pattern and skip anything else up to the closing ]
		if pattern[i+1] == '-' && i+3 < len(pattern) {
			if test >= c && test <= rune(pattern[i+2]) {
				result, offset := handleMatchInSet(negated, pattern[i+2:])
				return result, i + offset + 2
			}
		}

		// otherwise check for normal inclusion in the set
		if c == test {
			result, offset := handleMatchInSet(negated, pattern[i:])
			return result, i + offset
		}
	}

	// we didn't find the closing bracket
	return false, 0
}

func handleMatchInSet(negated bool, pattern string) (bool, int) {

	// if we matched when we were supposed to be negated, that's a failed match
	if negated {
		return false, 0
	}

	// ensure the closing bracket exists,
	// and return the offset of the next byte beyond it
	idx := strings.IndexRune(pattern, ']')
	return idx >= 0, idx + 1
}
