package osc

import (
	// "fmt"
	"testing"
)

func TestBasics(t *testing.T) {

	// empty matchers
	if !PatternMatch("", "") {
		t.Error()
	}

	if PatternMatch("a", "") {
		t.Error()
	}

	if PatternMatch("", "a") {
		t.Error()
	}

	// patterns that only diverge at the last character
	if PatternMatch("/blah2", "/blah") {
		t.Error()
	}

	if PatternMatch("/blah", "/blah2") {
		t.Error()
	}

	// normal match
	if !PatternMatch("/an/example", "/an/example") {
		t.Error()
	}
}

func TestSingleCharWildcard(t *testing.T) {

	if !PatternMatch("/a?c", "/abc") {
		t.Error()
	}

	if !PatternMatch("/a?c", "/a9c") {
		t.Error()
	}

	if !PatternMatch("/a?c/x??", "/aSc/xPG") {
		t.Error()
	}
}

func TestEscapedCharacter(t *testing.T) {

	// special character is properly escaped and matches the corresponding char in the test
	if !PatternMatch("/a\\*c", "/a*c") {
		t.Error()
	}

	// character after escape char doesn't match
	if PatternMatch("/a\\]", "/a[") {
		t.Error()
	}

	// escape char as last in message - matches empty space
	if !PatternMatch("/a\\", "/a") {
		t.Error()
	}

	// but not anything else
	if PatternMatch("/a\\", "/ag") {
		t.Error()
	}
}

func TestWildcard(t *testing.T) {

	if !PatternMatch("/a/*/c", "/a/9sdfhsdgh/c") {
		t.Error()
	}

	// test as last element
	if !PatternMatch("/a/*", "/a/9sdfhsdgh") {
		t.Error()
	}

	// different number of message parts
	if PatternMatch("/a/*", "/a/9sdfhsdgh/c") {
		t.Error()
	}

	if PatternMatch("/a/*/z", "/a/9sdfhsdgh") {
		t.Error()
	}

	// somewhat silly pattern string - initial * trumps everything to the next delimiter
	if !PatternMatch("/a/*other/z", "/a/bingo/z") {
		t.Error()
	}

	// wildcard partial way through a part
	if !PatternMatch("/a/prefix*/z", "/a/prefixsuffix/z") {
		t.Error()
	}
}

func TestSet(t *testing.T) {

	// set range is inclusive
	// make sure we catch either extreme and a couple instances in between
	if !PatternMatch("/[a-z]", "/a") {
		t.Error()
	}

	if !PatternMatch("/[a-z]", "/z") {
		t.Error()
	}

	if !PatternMatch("/[a-z]", "/m") {
		t.Error()
	}

	// non-range mode
	if !PatternMatch("/[abc]", "/b") {
		t.Error()
	}

	if !PatternMatch("/[abc]", "/c") {
		t.Error()
	}

	if PatternMatch("/[abc]", "/d") {
		t.Error()
	}

	if PatternMatch("/[abc]", "/d") {
		t.Error()
	}

	// at an offset into the pattern
	if !PatternMatch("/a[xyz]c", "/ayc") {
		t.Error()
	}

	// negated
	if !PatternMatch("/[!abc]", "/d") {
		t.Error()
	}
}
