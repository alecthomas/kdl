package kdl

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/repr"
	"github.com/stretchr/testify/require"
)

func TestFixtures(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/*.kdl")
	require.NoError(t, err)
	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			r, err := os.Open(fixture)
			require.NoError(t, err)
			defer r.Close() // nolint
			doc, err := Parse(fixture, r)
			require.NoError(t, err, repr.String(doc, repr.Indent("  ")))
			repr.Println(doc, repr.Indent("  "))
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Document
	}{
		{`BareNode`, `node`,
			doc(node("node"))},
		{`ArgNumber`, `value 123`,
			doc(node("value", arg(number(123))))},
		{`ArgString`, `value "abc"`,
			doc(node("value", arg(str("abc"))))},
		{`ArgRawString`, `value r"abc"`,
			doc(node("value", arg(str("abc"))))},
		{`ArgRawStringWithDelimiter`, `value r##"abc"##`,
			doc(node("value", arg(str("abc"))))},
		{`ArgBool`, `value true`,
			doc(node("value", arg(boolean(true))))},
		{`ArgBool`, `value false`,
			doc(node("value", arg(boolean(false))))},
		{`ArgNull`, `value null`,
			doc(node("value", arg(null())))},
		{`ArgMultiple`, `node 123 "arg"`,
			doc(node("node", arg(number(123)), arg(str("arg"))))},
		{`PropNumber`, `value num=123`,
			doc(node("value", prop("num", number(123))))},
		{`PropString`, `value str="abc"`,
			doc(node("value", prop("str", str("abc"))))},
		{`PropRawString`, `value str=r"abc"`,
			doc(node("value", prop("str", str("abc"))))},
		{`PropRawStringWithDelimiter`, `value str=r##"abc"##`,
			doc(node("value", prop("str", str("abc"))))},
		{`PropBool`, `value bool=true`,
			doc(node("value", prop("bool", boolean(true))))},
		{`PropNull`, `value n=null`,
			doc(node("value", prop("n", null())))},
		{`PropMultiple`, `node num=123 str="prop"`,
			doc(node("node", prop("num", number(123)), prop("str", str("prop"))))},
		{`ArgPropInterspersed`, `node 123 prop="abc"`,
			doc(node("node", arg(number(123)), prop("prop", str("abc"))))},
		{`InlineBlock`, `outer { inner }`,
			doc(block(node("outer"), node("inner")))},
		{`MultilineBlock`, `
			outer {
				first prop="abc"
				second
			}`,
			doc(block(node("outer"),
				node("first", prop("prop", str("abc"))),
				node("second")))},
		{`NodeArgContinuation`, `
			node \
				"arg"
				`, doc(node("node", arg(str("arg"))))},
		{`MultilineString`, `
			node "arg
long"
			`, doc(node("node", arg(str("arg\nlong"))))},
		{`MultilineRawString`, `
			node r"arg
long"
			`, doc(node("node", arg(str("arg\nlong"))))},
		{`StringWithEscapes`, `
			node "arg\nlong"
			`, doc(node("node", arg(str("arg\nlong"))))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseString("", test.input)
			require.NoError(t, err, repr.String(actual, repr.Indent("  ")))
			require.Equal(t,
				repr.String(test.expected, repr.Indent("  ")),
				repr.String(actual, repr.Indent("  ")))
		})
	}
}

func doc(nodes ...*Node) *Document {
	return &Document{Nodes: nodes}
}

func node(name string, params ...*Parameter) *Node {
	return &Node{Name: name, Parameters: params}
}

func block(root *Node, entries ...*Node) *Node {
	root.Block = &Block{Nodes: entries}
	return root
}

func arg(v *Value) *Parameter {
	return &Parameter{Argument: v}
}

func prop(name string, v *Value) *Parameter {
	return &Parameter{Property: &Property{
		Name:  name,
		Value: v,
	}}
}

func number(n float64) *Value { // nolint: unparam
	// Emulate what big.Float.UnmarshalText does.
	f := &big.Float{}
	_, _, _ = f.Parse(fmt.Sprintf("%v", n), 0)
	return &Value{Number: f}
}

func str(s string) *Value {
	return &Value{String: &s}
}

func boolean(v bool) *Value {
	return &Value{Bool: (*Bool)(&v)}
}

func null() *Value {
	return &Value{Null: true}
}
