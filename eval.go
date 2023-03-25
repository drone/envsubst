package envsubst

import "os"

// Eval replaces ${var} in the string based on the mapping function.
func Eval(s string, mapping func(string) (string, bool)) (string, error) {
	t, err := Parse(s)
	if err != nil {
		return s, err
	}
	return t.Execute(mapping)
}

// EvalEnv replaces ${var} in the string according to the values of the
// current environment variables. References to undefined variables are
// replaced by the empty string.
func EvalEnv(s string, strict bool) (string, error) {
	mapping := Getenv
	if strict{
		mapping = os.LookupEnv
	}
	return Eval(s, mapping)
}

func Getenv(s string) (string, bool) {
	return os.Getenv(s), true
}