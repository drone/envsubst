package parse

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	// ErrBadSubstitution represents a substitution parsing error.
	ErrBadSubstitution = errors.New("bad substitution")
	ErrBadSubstitution2 = errors.New("bad substitution")

	// ErrMissingClosingBrace represents a missing closing brace "}" error.
	ErrMissingClosingBrace = errors.New("missing closing brace")

	// ErrParseVariableName represents the error when unable to parse a
	// variable name within a substitution.
	ErrParseVariableName = errors.New("unable to parse variable name")

	// ErrParseFuncSubstitution represents the error when unable to parse the
	// substitution within a function parameter.
	ErrParseFuncSubstitution = errors.New("unable to parse substitution within function")

	// ErrParseDefaultFunction represent the error when unable to parse a
	// default function.
	ErrParseDefaultFunction = errors.New("unable to parse default function")
)

// ErrParseDoubleDollar represents the error when unable to parse a $$
func ErrParseDoubleDollar(str string) error {
	return fmt.Errorf("unable to parse double dollar sign %s", str)
}

// Tree is the representation of a single parsed shell format string
type Tree struct {
	Root Node

	// Parsing only; cleared after parse.
	scanner *scanner
}

// Parse parses the string and returns a Tree.
func Parse(buf string) (*Tree, error) {
	t := new(Tree)
	t.scanner = new(scanner)
	return t.Parse(buf)
}

// Parse parses the string buffer to construct an ast
// representation for expansion.
func (t *Tree) Parse(buf string) (tree *Tree, err error) {
	t.scanner.init(buf)
	t.Root, err = t.parseAny()
	return t, err
}

func (t *Tree) parseAny() (Node, error) {
	t.scanner.accept = acceptRune
	t.scanner.mode = scanIdent | scanLbrack | scanEscape
	t.scanner.escapeChars = dollar

	switch t.scanner.scan() {
	case tokenIdent:
		left := newTextNode(
			t.scanner.string(),
		)
		right, err := t.parseAny()
		switch {
		case err != nil:
			return nil, err
		case right == empty:
			return left, nil
		}
		return newListNode(left, right), nil
	case tokenEOF:
		return empty, nil
	case tokenLbrack:
		left, err := t.parseFunc()
		if err != nil {
			return nil, err
		}

		right, err := t.parseAny()
		switch {
		case err != nil:
			return nil, err
		case right == empty:
			return left, nil
		}
		return newListNode(left, right), nil
	case tokenBarevar:
		left, err := t.parseBareVar()
		if err != nil {
			return nil, err
		}

		right, err := t.parseAny()
		switch {
		case err != nil:
			return nil, err
		case right == empty:
			return left, nil
		}
		return newListNode(left, right), nil
	case tokenDoubleDollar:
		left := newTextNode("$")

		right, err := t.parseAny()
		switch {
		case err != nil:
			return nil, err
		case right == empty:
			return left, nil
		}
		return newListNode(left, right), nil
	}

	return nil, ErrBadSubstitution
}

func (t *Tree) parseBareVar() (Node, error) {
	t.scanner.accept = acceptIdent
	t.scanner.mode = scanIdent

	var name string
	switch t.scanner.scan() {
	case tokenIdent:
		name = t.scanner.string()
	default:
		return nil, ErrParseVariableName
	}

	node := newFuncNode(name)
	_, err := node.buf.Write([]byte("$" + name))
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (t *Tree) parseFunc() (Node, error) {
	// Turn on all escape characters
	t.scanner.escapeChars = escapeAll
	switch t.scanner.peek() {
	case '#':
		return t.parseLenFunc()
	}

	var name string
	t.scanner.accept = acceptIdent
	t.scanner.mode = scanIdent

	switch t.scanner.scan() {
	case tokenIdent:
		name = t.scanner.string()
	default:
		return nil, ErrParseVariableName
	}

	switch t.scanner.peek() {
	case ':':
		return t.parseDefaultOrSubstr(name)
	case '=':
		return t.parseDefaultFunc(name)
	case ',', '^':
		return t.parseCasingFunc(name)
	case '/':
		return t.parseReplaceFunc(name)
	case '#':
		return t.parseRemoveFunc(name, acceptHashFunc)
	case '%':
		return t.parseRemoveFunc(name, acceptPercentFunc)
	}

	// trivial case: ${var}

	t.scanner.accept = acceptIdent
	t.scanner.mode = scanRbrack | scanIdent | scanLbrack | scanEscape
	switch t.scanner.scan() {
	case tokenRbrack:
		node := newFuncNode(name)
		_, err := node.buf.Write([]byte("${" + name + "}"))
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		return nil, ErrMissingClosingBrace
	}
}

// parse a substitution function parameter.
func (t *Tree) parseParam(accept acceptFunc, mode byte) (Node, error) {
	t.scanner.accept = accept
	t.scanner.mode = mode | scanLbrack
	switch t.scanner.scan() {
	case tokenLbrack:
		return t.parseFunc()
	case tokenBarevar:
		return t.parseBareVar()
	case tokenDoubleDollar:
		left := newTextNode("$")

		right, err := t.parseParam(accept, mode)
		switch {
		case err != nil:
			return nil, err
		case right == empty:
			return left, nil
		}
		return newListNode(left, right), nil
	case tokenIdent:
		// TODO maybe add a } here?
		return newTextNode(
			t.scanner.string(),
		), nil
	case tokenRbrack:
		return newTextNode(
			t.scanner.string(),
		), nil
	default:
		return nil, ErrParseFuncSubstitution
	}
}

// parse either a default or substring substitution function.
func (t *Tree) parseDefaultOrSubstr(name string) (Node, error) {
	// selects between default or substr
	// no need for writing original string to node yet

	switch t.scanner.peektwo() {
	case '=', '-', '?', '+':
		return t.parseDefaultFunc(name)
	default:
		return t.parseSubstrFunc(name)
	}
}

type nodeFormatter struct {
	buf bytes.Buffer
}

func (f *nodeFormatter) getFormat(node Node) {
	switch n := node.(type) {
	case *TextNode:
		f.buf.WriteString(n.Value)
	case *ListNode:
		for _, item := range n.Nodes {
			f.buf.WriteString(FormatNode(item))
		}
	case *FuncNode:
		f.buf.WriteString(n.String())
	}
}

func FormatNode(node Node) string {
	f := new(nodeFormatter)
	f.getFormat(node)
	return f.buf.String()
}

// parses the ${param:offset} string function
// parses the ${param:offset:length} string function
func (t *Tree) parseSubstrFunc(name string) (Node, error) {
	node := new(FuncNode)
	node.Param = name
	_, err := node.buf.WriteString("${" + name)
	if err != nil {
		return nil, err
	}

	t.scanner.accept = acceptOneColon
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeName := t.scanner.string()
		node.Name = nodeName
		_, err := node.buf.WriteString(nodeName)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	// scan arg[1]
	{
		param, err := t.parseParam(rejectColonClose, scanIdent)
		if err != nil {
			return nil, err
		}

		_, err = node.buf.WriteString(FormatNode(param))
		if err != nil {
			return nil, err
		}

		switch n := param.(type) {
		case *FuncNode:
			n.nesting = node.nesting + 1

			node.Args = append(node.Args, n)
		default:
			node.Args = append(node.Args, param)
		}
	}

	// expect delimiter or close
	t.scanner.accept = acceptColon
	t.scanner.mode = scanIdent | scanRbrack
	switch t.scanner.scan() {
	case tokenRbrack:
		t.scanner.unread()
		return node, t.consumeRbrack(node)
	case tokenIdent:
		delimiter := t.scanner.string()
		_, err := node.buf.WriteString(delimiter)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	// scan arg[2]
	{
		param, err := t.parseParam(acceptNotClosing, scanIdent)
		if err != nil {
			return nil, err
		}

		_, err = node.buf.WriteString(FormatNode(param))
		if err != nil {
			return nil, err
		}

		switch n := param.(type) {
		case *FuncNode:
			n.nesting = node.nesting + 1

			node.Args = append(node.Args, n)
		default:
			node.Args = append(node.Args, param)
		}
	}

	return node, t.consumeRbrack(node)
}

// parses the ${param%word} string function
// parses the ${param%%word} string function
// parses the ${param#word} string function
// parses the ${param##word} string function
func (t *Tree) parseRemoveFunc(name string, accept acceptFunc) (Node, error) {
	node := new(FuncNode)
	node.Param = name
	_, err := node.buf.WriteString("${" + name)
	if err != nil {
		return nil, err
	}

	t.scanner.accept = accept
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeName := t.scanner.string()
		node.Name = nodeName
		_, err := node.buf.WriteString(nodeName)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	// scan arg[1]
	{
		param, err := t.parseParam(acceptNotClosing, scanIdent)
		if err != nil {
			return nil, err
		}

		_, err = node.buf.WriteString(FormatNode(param))
		if err != nil {
			return nil, err
		}

		switch n := param.(type) {
		case *FuncNode:
			n.nesting = node.nesting + 1

			node.Args = append(node.Args, n)
		default:
			node.Args = append(node.Args, param)
		}
	}

	return node, t.consumeRbrack(node)
}

// parses the ${param/pattern/string} string function
// parses the ${param//pattern/string} string function
// parses the ${param/#pattern/string} string function
// parses the ${param/%pattern/string} string function
func (t *Tree) parseReplaceFunc(name string) (Node, error) {
	node := new(FuncNode)
	node.Param = name
	_, err := node.buf.WriteString("${" + name)
	if err != nil {
		return nil, err
	}

	t.scanner.accept = acceptReplaceFunc
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeName := t.scanner.string()
		node.Name = nodeName
		_, err := node.buf.WriteString(nodeName)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	// scan arg[1]
	{
		param, err := t.parseParam(acceptNotSlash, scanIdent|scanEscape)
		if err != nil {
			return nil, err
		}

		_, err = node.buf.WriteString(FormatNode(param))
		if err != nil {
			return nil, err
		}

		switch n := param.(type) {
		case *FuncNode:
			n.nesting = node.nesting + 1

			node.Args = append(node.Args, n)
		default:
			node.Args = append(node.Args, param)
		}
	}

	// expect delimiter or close
	t.scanner.accept = acceptSlash
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenRbrack:
		return node, t.consumeRbrack(node)
	case tokenIdent:
		delimiter := t.scanner.string()
		_, err := node.buf.WriteString(delimiter)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	// scan arg[2]
	{
		param, err := t.parseParam(acceptNotClosing, scanIdent|scanEscape)
		if err != nil {
			return nil, err
		}

		_, err = node.buf.WriteString(FormatNode(param))
		if err != nil {
			return nil, err
		}

		switch n := param.(type) {
		case *FuncNode:
			n.nesting = node.nesting + 1

			node.Args = append(node.Args, n)
		default:
			node.Args = append(node.Args, param)
		}
	}

	return node, t.consumeRbrack(node)
}

// parses the ${parameter=word} string function
// parses the ${parameter:=word} string function
// parses the ${parameter:-word} string function
// parses the ${parameter:?word} string function
// parses the ${parameter:+word} string function
func (t *Tree) parseDefaultFunc(name string) (Node, error) {
	node := new(FuncNode)
	node.Param = name
	_, err := node.buf.WriteString("${" + name)
	if err != nil {
		return nil, err
	}

	t.scanner.accept = acceptDefaultFunc
	if t.scanner.peek() == '=' {
		t.scanner.accept = acceptOneEqual
	}
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeName := t.scanner.string()
		node.Name = nodeName
		_, err := node.buf.WriteString(nodeName)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrParseDefaultFunction
	}

	// loop through all possible runes in default param
	for {
		// this acts as the break condition. Peek to see if we reached the end
		switch t.scanner.peek() {
		case '}':
			return node, t.consumeRbrack(node)
		}
		param, err := t.parseParam(acceptNotClosing, scanIdent)
		if err != nil {
			return nil, err
		}

		_, err = node.buf.WriteString(FormatNode(param))
		if err != nil {
			return nil, err
		}

		switch n := param.(type) {
		case *FuncNode:
			n.nesting = node.nesting + 1

			node.Args = append(node.Args, n)
		default:
			node.Args = append(node.Args, param)
		}
	}
}

// parses the ${param,} string function
// parses the ${param,,} string function
// parses the ${param^} string function
// parses the ${param^^} string function
func (t *Tree) parseCasingFunc(name string) (Node, error) {
	node := new(FuncNode)
	node.Param = name
	_, err := node.buf.WriteString("${" + name)
	if err != nil {
		return nil, err
	}

	t.scanner.accept = acceptCasingFunc
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeName := t.scanner.string()
		node.Name = nodeName
		_, err := node.buf.WriteString(nodeName)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	return node, t.consumeRbrack(node)
}

// parses the ${#param} string function
func (t *Tree) parseLenFunc() (Node, error) {
	node := new(FuncNode)
	_, err := node.buf.WriteString("${")
	if err != nil {
		return nil, err
	}

	t.scanner.accept = acceptOneHash
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeName := t.scanner.string()
		node.Name = nodeName
		_, err := node.buf.WriteString(nodeName)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	t.scanner.accept = acceptIdent
	t.scanner.mode = scanIdent
	switch t.scanner.scan() {
	case tokenIdent:
		nodeParam := t.scanner.string()
		node.Param = nodeParam
		_, err := node.buf.WriteString(nodeParam)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrBadSubstitution
	}

	return node, t.consumeRbrack(node)
}

// consumeRbrack consumes a right closing bracket. If a closing
// bracket token is not consumed an ErrBadSubstitution is returned.
func (t *Tree) consumeRbrack(node *FuncNode) error {
	t.scanner.mode = scanRbrack
	if t.scanner.scan() != tokenRbrack {
		return ErrBadSubstitution
	}
	_, err := node.buf.Write([]byte("}"))

	return err
}
