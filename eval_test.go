package envsubst

import (
	"fmt"
	"testing"
)

// test cases sourced from tldp.org
// http://www.tldp.org/LDP/abs/html/parameter-substitution.html

func TestExpand(t *testing.T) {
	var expressions = []struct {
		params map[string]string
		input  string
		output string
		err error
	}{
		// text-only
		{
			params: map[string]string{},
			input:  "abcdEFGH28ij",
			output: "abcdEFGH28ij",
		},
		// length
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${#var01}",
			output: "12",
		},
		// uppercase first
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${var01^}",
			output: "AbcdEFGH28ij",
		},
		// uppercase
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${var01^^}",
			output: "ABCDEFGH28IJ",
		},
		// lowercase first
		{
			params: map[string]string{"var01": "ABCDEFGH28IJ"},
			input:  "${var01,}",
			output: "aBCDEFGH28IJ",
		},
		// lowercase
		{
			params: map[string]string{"var01": "ABCDEFGH28IJ"},
			input:  "${var01,,}",
			output: "abcdefgh28ij",
		},
		// substring with position
		{
			params: map[string]string{"path_name": "/home/bozo/ideas/thoughts.for.today"},
			input:  "${path_name:11}",
			output: "ideas/thoughts.for.today",
		},
		// substring with position and length
		{
			params: map[string]string{"path_name": "/home/bozo/ideas/thoughts.for.today"},
			input:  "${path_name:11:5}",
			output: "ideas",
		},
		// default not used
		{
			params: map[string]string{"var": "abc"},
			input:  "${var=xyz}",
			output: "abc",
		},
		// default used
		{
			params: map[string]string{},
			input:  "${var=xyz}",
			output: "xyz",
		},
		{
			params: map[string]string{"default_var": "foo"},
			input:  "something ${var=${default_var}}",
			output: "something foo",
		},
		{
			params: map[string]string{"default_var": "foo1"},
			input:  `foo: ${var=${default_var}-suffix}`,
			output: "foo: foo1-suffix",
		},
		{
			params: map[string]string{"default_var": "foo1"},
			input:  `foo: ${var=prefix${default_var}-suffix}`,
			output: "foo: prefixfoo1-suffix",
		},
		{
			params: map[string]string{},
			input:  "${var:=xyz}",
			output: "xyz",
		},
		// replace suffix
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ/%abc/XYZ}",
			output: "abcABC123ABCXYZ",
		},
		// replace prefix
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ/#abc/XYZ}",
			output: "XYZABC123ABCabc",
		},
		// replace all
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ//abc/xyz}",
			output: "xyzABC123ABCxyz",
		},
		// replace first
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ/abc/xyz}",
			output: "xyzABC123ABCabc",
		},
		// delete shortest match prefix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename#*.}",
			output: "string.txt",
		},
		{
			params: map[string]string{"filename": "path/to/file"},
			input:  "${filename#*/}",
			output: "to/file",
		},
		{
			params: map[string]string{"filename": "/path/to/file"},
			input:  "${filename#*/}",
			output: "path/to/file",
		},
		// delete longest match prefix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename##*.}",
			output: "txt",
		},
		{
			params: map[string]string{"filename": "path/to/file"},
			input:  "${filename##*/}",
			output: "file",
		},
		{
			params: map[string]string{"filename": "/path/to/file"},
			input:  "${filename##*/}",
			output: "file",
		},
		// delete shortest match suffix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename%.*}",
			output: "bash.string",
		},
		// delete longest match suffix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename%%.*}",
			output: "bash",
		},

		// nested parameters
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${var=${var01^^}}",
			output: "ABCDEFGH28IJ",
		},
		// escaped
		{
			params: map[string]string{"var": "foo"},
			input:  "some text ${var}$${var$${var}$var${var}",
			output: "some text foo$${var$foofoofoo",
			err:  fmt.Errorf("missing closing brace"),
		},
		{
			params: map[string]string{"default_var": "foo"},
			input:  "something $${var=${default_var}}",
			output: "something $foo",
		},
		// some common escaping use cases
		{
			params: map[string]string{"stringZ": "foo/bar"},
			input:  `${stringZ/\//-}`,
			output: "foo-bar",
		},
		{
			params: map[string]string{"stringZ": "foo/bar/baz"},
			input:  `${stringZ//\//-}`,
			output: "foo-bar-baz",
		},
		// escape outside of expansion shouldn't be processed
		{
			params: map[string]string{"default_var": "foo"},
			input:  "\\\\something ${var=${default_var}}",
			output: "\\\\something foo",
		},
		// substitute with a blank string
		{
			params: map[string]string{"stringZ": "foo.bar"},
			input:  `${stringZ/./}`,
			output: "foobar",
		},
		// dollar sign handling
		{
			params: map[string]string{"": ""},
			input:  "$",
			output: "$",
		},
		{
			params: map[string]string{"": ""},
			input:  "$$",
			output: "$$",
		},
		{
			params: map[string]string{"": ""},
			input:  "$$$",
			output: "$$$",
		},
		{
			params: map[string]string{"var": "foo"},
			input:  "$$var",
			output: "$foo",
		},
		{
			params: map[string]string{"var": "foo"},
			input:  "$${var}",
			output: "$foo",
		},
		// nested
		{
			params: map[string]string{"var": "foo"},
			input:  "${var2:-$var}",
			output: "foo",
		},
		{
			params: map[string]string{"var": "foo"},
			input:  "${var2:-${var}}",
			output: "foo",
		},
		{
			params: map[string]string{"": ""},
			input:  "${var:-$$}",
			output: "$$",
		},
		{
			params: map[string]string{"var": "foo"},
			input:  `${var2:-"${var}lala"}`,
			output: `"foolala"`,
		},
		{
			params: map[string]string{"var": "foo"},
			input:  `${var:-$$}`,
			output: `foo`,
		},
		{
			params: map[string]string{"": ""},
			input:  `${var:-$$}`,
			output: `$$`,
		},
		{
			params: map[string]string{"": ""},
			input:  `${var:-${var2:-$$}}`,
			output: `$$`,
		},
		// newline
		{
			params: map[string]string{"": ""},
			input:  `${var:-${var2:-$$}}`,
			output: `$$`,
		},
	}

	for _, expr := range expressions {
		t.Run(expr.input, func(t *testing.T) {
			t.Logf(expr.input)
			output, err := Eval(expr.input, func(s string) string {
				return expr.params[s]
			})
			if expr.err != nil {
				if expr.err.Error() != err.Error() {
					t.Fatalf("Want error: %q expanded but got error: %q", expr.err.Error(), err)
				}
				// got expected error. go next test case
				return
			}
			if err != nil {
				t.Fatalf("Want %q expanded but got error %q", expr.input, err)
			}

			if output != expr.output {
				t.Fatalf("Want %q expanded to %q, got %q",
					expr.input,
					expr.output,
					output)
			}
		})
	}
}
