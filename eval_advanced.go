package envsubst

import (
	"bytes"
	"io"

	"github.com/logandavies181/envsubst/parse"
)

type NodeInfo struct {
	node parse.Node
}

// Orig returns the original text of the substitution template,
// before it was parsed. This can be used to provide full context
// for custom mapping functions or leave expressions un-evaluated
func (n NodeInfo) Orig() string {
	return parse.FormatNode(n.node)
}

// AdvancedMapping is a function that takes a variable name and
// representation of the full shell variable string and returns the substituted
// string and whether or not to continue processing
type AdvancedMapping func(string, NodeInfo) (string, bool)

// EvalAdvanced allows the caller to control how ${var} is mapped and how its
// nested parameters are evaluated
func EvalAdvanced(s string, mapping AdvancedMapping) (string, error) {
	t, err := Parse(s)
	if err != nil {
		return s, err
	}
	return t.ExecuteAdvanced(mapping)
}

// ExecuteAdvanced applies a parsed template to the specified data mapping,
// allowing greater control over execution
func (t *Template) ExecuteAdvanced(mapping AdvancedMapping) (str string, err error) {
	b := new(bytes.Buffer)
	s := new(state)
	s.node = t.tree.Root
	s.advMapper = mapping
	s.writer = b
	err = t.evalAdvanced(s)
	if err != nil {
		return
	}
	return "", nil
}

func (t *Template) evalAdvanced(s *state) (err error) {
	switch node := s.node.(type) {
	case *parse.TextNode:
		err = t.evalText(s, node)
	case *parse.FuncNode:
		err = t.evalAdvancedFunc(s, node)
	case *parse.ListNode:
		err = t.evalAdvancedList(s, node)
	}
	return err
}

func (t *Template) evalAdvancedList(s *state, node *parse.ListNode) (err error) {
	for _, n := range node.Nodes {
		s.node = n
		err = t.evalAdvanced(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) evalAdvancedFunc(s *state, node *parse.FuncNode) error {
	var w = s.writer
	var buf bytes.Buffer
	var args []string

	// restore the origin writer
	s.writer = w
	s.node = node

	val, shouldContinue := s.advMapper(node.Param, NodeInfo{node})

	if !shouldContinue {
		_, err := io.WriteString(s.writer, val)
		return err
	}

	for _, n := range node.Args {
		buf.Reset()
		s.writer = &buf
		s.node = n
		err := t.evalAdvanced(s)
		if err != nil {
			return err
		}
		args = append(args, buf.String())
	}
	fn := lookupFunc(node.Name, len(args))

	_, err := io.WriteString(s.writer, fn(val, args...))
	return err
}
