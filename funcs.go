package envsubst

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/woodpecker-ci/envsubst/path"
)

var (
	// ErrBadSubstringPosition represents a substring position that is
	// not an integer.
	ErrBadSubstringPosition = errors.New("bad substring position")

	// ErrBadSubstringLength represents a substring length that is not
	// an integer or, if negative, ends before the position.
	ErrBadSubstringLength = errors.New("bad substring length")

	// ErrMissingArgument represents a substitution function called
	// without the arguments it needs.
	ErrMissingArgument = errors.New("missing argument")
)

// defines a parameter substitution function.
type substituteFunc func(string, ...string) (string, error)

// toLen returns the length of string s.
func toLen(s string, args ...string) (string, error) {
	return strconv.Itoa(len(s)), nil
}

// toLower returns a copy of the string s with all characters
// mapped to their lower case.
func toLower(s string, args ...string) (string, error) {
	return strings.ToLower(s), nil
}

// toUpper returns a copy of the string s with all characters
// mapped to their upper case.
func toUpper(s string, args ...string) (string, error) {
	return strings.ToUpper(s), nil
}

// toLowerFirst returns a copy of the string s with the first
// character mapped to its lower case.
func toLowerFirst(s string, args ...string) (string, error) {
	if s == "" {
		return s, nil
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:], nil
}

// toUpperFirst returns a copy of the string s with the first
// character mapped to its upper case.
func toUpperFirst(s string, args ...string) (string, error) {
	if s == "" {
		return s, nil
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:], nil
}

// toDefault returns a copy of the string s if not empty, else
// returns a concatenation of the args without a separator.
func toDefault(s string, args ...string) (string, error) {
	if len(s) == 0 && len(args) > 0 {
		// don't use any separator
		s = strings.Join(args, "")
	}
	return s, nil
}

// toSubstr returns a slice of the string s at the specified
// length and position.
func toSubstr(s string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", ErrMissingArgument
	}

	pos, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrBadSubstringPosition, args[0])
	}

	if pos < 0 {
		// if pos is negative (counts from the end) add it
		// to length to get first character offset
		pos = len(s) + pos

		// if negative offset exceeds the length of the string
		// start from 0
		if pos < 0 {
			pos = 0
		}
	}

	if len(args) == 1 {
		if pos < len(s) {
			return s[pos:], nil
		}
		// if the position exceeds the length of the
		// string an empty string is returned
		return "", nil
	}

	length, err := strconv.Atoi(args[1])
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrBadSubstringLength, args[1])
	}

	if length < 0 {
		// like bash a negative length counts back from the
		// end of the string
		if pos > len(s) {
			return "", nil
		}
		length = len(s) + length - pos
		if length < 0 {
			// bash fails with "substring expression < 0"
			return "", fmt.Errorf("%w: %q", ErrBadSubstringLength, args[1])
		}
	}

	if pos+length >= len(s) {
		if pos < len(s) {
			// if the position exceeds the length of the
			// string just return the rest of it like bash
			return s[pos:], nil
		}
		// if the position exceeds the length of the
		// string an empty string is returned
		return "", nil
	}

	return s[pos : pos+length], nil
}

// replaceAll returns a copy of the string s with all instances
// of the substring replaced with the replacement string.
func replaceAll(s string, args ...string) (string, error) {
	switch len(args) {
	case 0:
		return "", ErrMissingArgument
	case 1:
		return strings.Replace(s, args[0], "", -1), nil
	default:
		return strings.Replace(s, args[0], args[1], -1), nil
	}
}

// replaceFirst returns a copy of the string s with the first
// instance of the substring replaced with the replacement string.
func replaceFirst(s string, args ...string) (string, error) {
	switch len(args) {
	case 0:
		return "", ErrMissingArgument
	case 1:
		return strings.Replace(s, args[0], "", 1), nil
	default:
		return strings.Replace(s, args[0], args[1], 1), nil
	}
}

// replacePrefix returns a copy of the string s with the matching
// prefix replaced with the replacement string.
func replacePrefix(s string, args ...string) (string, error) {
	if len(args) != 2 {
		return "", ErrMissingArgument
	}
	if strings.HasPrefix(s, args[0]) {
		return strings.Replace(s, args[0], args[1], 1), nil
	}
	return s, nil
}

// replaceSuffix returns a copy of the string s with the matching
// suffix replaced with the replacement string.
func replaceSuffix(s string, args ...string) (string, error) {
	if len(args) != 2 {
		return "", ErrMissingArgument
	}
	if strings.HasSuffix(s, args[0]) {
		s = strings.TrimSuffix(s, args[0])
		s = s + args[1]
	}
	return s, nil
}

// TODO

func trimShortestPrefix(s string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", ErrMissingArgument
	}
	return trimShortest(s, args[0])
}

func trimShortestSuffix(s string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", ErrMissingArgument
	}
	r := reverse(s)
	rarg := reverse(args[0])
	s, err := trimShortest(r, rarg)
	return reverse(s), err
}

func trimLongestPrefix(s string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", ErrMissingArgument
	}
	return trimLongest(s, args[0])
}

func trimLongestSuffix(s string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", ErrMissingArgument
	}
	r := reverse(s)
	rarg := reverse(args[0])
	s, err := trimLongest(r, rarg)
	return reverse(s), err
}

func trimShortest(s, arg string) (string, error) {
	var shortestMatch string
	for i := 0; i < len(s); i++ {
		match, err := path.Match(arg, s[0:len(s)-i])
		if err != nil {
			return "", err
		}

		if match {
			shortestMatch = s[0 : len(s)-i]
		}
	}

	if shortestMatch != "" {
		return strings.TrimPrefix(s, shortestMatch), nil
	}

	return s, nil
}

func trimLongest(s, arg string) (string, error) {
	for i := 0; i < len(s); i++ {
		match, err := path.Match(arg, s[0:len(s)-i])
		if err != nil {
			return "", err
		}

		if match {
			return strings.TrimPrefix(s, s[0:len(s)-i]), nil
		}
	}

	return s, nil
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
