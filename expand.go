package envexp

import "os"

// Expand replaces ${var} in the string based on the mapping function.
func Expand(s string, mapping func(string) string) (string, error) {
	t, err := Parse(s)
	if err != nil {
		return s, err
	}
	return t.Execute(mapping)
}

// ExpandEnv replaces ${var} in the string according to the values of the
// current environment variables. References to undefined variables are
// replaced by the empty string.
func ExpandEnv(s string) (string, error) {
	return Expand(s, os.Getenv)
}
