package parse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var tests = []struct {
	Text string
	Node Node
}{

	//
	// text only
	//
	{
		Text: "text",
		Node: &TextNode{Value: "text"},
	},
	{
		Text: "}text",
		Node: &TextNode{Value: "}text"},
	},
	{
		Text: "http://github.com",
		Node: &TextNode{Value: "http://github.com"}, // should not escape double slash
	},
	{
		Text: "$${string}",
		Node: &TextNode{Value: "${string}"}, // should not escape double dollar
	},
	{
		Text: "$$string",
		Node: &TextNode{Value: "$string"}, // should not escape double dollar
	},
	{
		Text: `\\.\pipe\pipename`,
		Node: &TextNode{Value: `\\.\pipe\pipename`},
	},

	//
	// variable only
	//
	{
		Text: "${string}",
		Node: &FuncNode{Param: "string"},
	},

	//
	// text transform functions
	//
	{
		Text: "${string,}",
		Node: &FuncNode{
			Param: "string",
			Name:  ",",
			Args:  nil,
		},
	},
	{
		Text: "${string,,}",
		Node: &FuncNode{
			Param: "string",
			Name:  ",,",
			Args:  nil,
		},
	},
	{
		Text: "${string^}",
		Node: &FuncNode{
			Param: "string",
			Name:  "^",
			Args:  nil,
		},
	},
	{
		Text: "${string^^}",
		Node: &FuncNode{
			Param: "string",
			Name:  "^^",
			Args:  nil,
		},
	},

	//
	// substring functions
	//
	{
		Text: "${string:position}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&TextNode{Value: "position"},
			},
		},
	},
	{
		Text: "${string:position:length}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&TextNode{Value: "position"},
				&TextNode{Value: "length"},
			},
		},
	},

	//
	// string removal functions
	//
	{
		Text: "${string#substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},
	{
		Text: "${string##substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "##",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},
	{
		Text: "${string%substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "%",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},
	{
		Text: "${string%%substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "%%",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},

	//
	// string replace functions
	//
	{
		Text: "${string/substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},
	{
		Text: "${string//substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "//",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},
	{
		Text: "${string/#substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "/#",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},
	{
		Text: "${string/%substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "/%",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},

	//
	// default value functions
	//
	{
		Text: "${string=default}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:=default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":=",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:-default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":-",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:?default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":?",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:+default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":+",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},

	//
	// length function
	//
	{
		Text: "${#string}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
		},
	},

	//
	// special characters in argument
	//
	{
		Text: "${string#$%:*{}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
			Args: []Node{
				&TextNode{Value: "$%:*{"},
			},
		},
	},

	// text before and after function
	{
		Text: "hello ${#string} world",
		Node: &ListNode{
			Nodes: []Node{
				&TextNode{
					Value: "hello ",
				},
				&ListNode{
					Nodes: []Node{
						&FuncNode{
							Param: "string",
							Name:  "#",
						},
						&TextNode{
							Value: " world",
						},
					},
				},
			},
		},
	},
	// text before and after function with \\ outside of function
	{
		Text: `\\ hello ${#string} world \\`,
		Node: &ListNode{
			Nodes: []Node{
				&TextNode{
					Value: `\\ hello `,
				},
				&ListNode{
					Nodes: []Node{
						&FuncNode{
							Param: "string",
							Name:  "#",
						},
						&TextNode{
							Value: ` world \\`,
						},
					},
				},
			},
		},
	},

	// escaped function arguments
	{
		Text: `${string/\/position/length}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "/position",
				},
				&TextNode{
					Value: "length",
				},
			},
		},
	},
	{
		Text: `${string/\/position\\/length}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "/position\\",
				},
				&TextNode{
					Value: "length",
				},
			},
		},
	},
	{
		Text: `${string/position/\/length}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "position",
				},
				&TextNode{
					Value: "/length",
				},
			},
		},
	},
	{
		Text: `${string/position/\/length\\}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "position",
				},
				&TextNode{
					Value: "/length\\",
				},
			},
		},
	},
	{
		Text: `${string/position/\/leng\\th}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "position",
				},
				&TextNode{
					Value: "/leng\\th",
				},
			},
		},
	},

	// functions in functions
	{
		Text: "${string:${position}}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&FuncNode{
					Param: "position",
				},
			},
		},
	},
	{
		Text: "${string:${stringy:position:length}:${stringz,,}}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&FuncNode{
					Param: "stringy",
					Name:  ":",
					Args: []Node{
						&TextNode{Value: "position"},
						&TextNode{Value: "length"},
					},
				},
				&FuncNode{
					Param: "stringz",
					Name:  ",,",
				},
			},
		},
	},
	{
		Text: "${string#${stringz}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
			Args: []Node{
				&FuncNode{Param: "stringz"},
			},
		},
	},
	{
		Text: "${string=${stringz}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&FuncNode{Param: "stringz"},
			},
		},
	},
	{
		Text: "${string=prefix-${var}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "prefix-"},
				&FuncNode{Param: "var"},
			},
		},
	},
	{
		Text: "${string=${var}-suffix}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&FuncNode{Param: "var"},
				&TextNode{Value: "-suffix"},
			},
		},
	},
	{
		Text: "${string=prefix-${var}-suffix}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "prefix-"},
				&FuncNode{Param: "var"},
				&TextNode{Value: "-suffix"},
			},
		},
	},
	{
		Text: "${string=prefix${var} suffix}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "prefix"},
				&FuncNode{Param: "var"},
				&TextNode{Value: " suffix"},
			},
		},
	},
	{
		Text: "${string//${stringy}/${stringz}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "//",
			Args: []Node{
				&FuncNode{Param: "stringy"},
				&FuncNode{Param: "stringz"},
			},
		},
	},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		t.Log(test.Text)
		t.Run(test.Text, func(t *testing.T) {
			got, err := Parse(test.Text)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Node, got.Root); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
