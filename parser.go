// Package kdl is parser for KDL (https://kdl.dev)
package kdl

import (
	"io"
	"math/big"

	"github.com/alecthomas/participle/v2"
)

var parser = participle.MustBuild(&Document{},
	participle.Lexer(fixupLexerDefinition{}),
	participle.Unquote(`String`),
	participle.UseLookahead(3),
)

// A Bool value.
type Bool bool

func (b *Bool) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

// Value is a scalar value.
type Value struct {
	String *string    `  (@String | (RawStringStart @RawStringText* RawStringEnd))`
	Number *big.Float `| @Number`
	Bool   *Bool      `| @("true" | "false")`
	Null   bool       `| @"null"`
}

type Property struct {
	Name  string `(@Ident | @String | (RawStringStart @RawStringText* RawStringEnd)) "="`
	Value *Value `@@`
}

// A Parameter is either a positional argument or a property.
type Parameter struct {
	Skip     bool      `@Skip?`
	Property *Property `  @@`
	Argument *Value    `| @@`
}

// A Block attached to a node.
type Block = Document

type Node struct {
	Skip       bool         `@Skip?`
	Name       string       `@(Ident | String)`
	Parameters []*Parameter `@@*`
	Block      *Block       `("{" @@ "}")? ";"+`
}

// Properties of the Node.
func (n *Node) Properties() map[string]*Value {
	out := make(map[string]*Value, len(n.Parameters))
	for _, param := range n.Parameters {
		if param.Property != nil {
			out[param.Property.Name] = param.Property.Value
		}
	}
	return out
}

type Document struct {
	Nodes []*Node `@@*`
}

// Parse a KDL document from "r".
func Parse(filename string, r io.Reader) (*Document, error) {
	doc := &Document{}
	return doc, parser.Parse(filename, r, doc)
}

// ParseString parses a KDL document from the provided string.
func ParseString(filename, text string) (*Document, error) {
	doc := &Document{}
	return doc, parser.ParseString(filename, text, doc)
}
