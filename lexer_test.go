package kdl

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/require"
)

func TestSemicolonInsertion(t *testing.T) {
	def := fixupLexerDefinition{}
	lex, err := def.Lex("", strings.NewReader(`
node0
node1 
block {
}
node2 \
arg
`))
	require.NoError(t, err)
	tokens, err := lexer.ConsumeAll(lex)
	require.NoError(t, err)
	actual := make([]string, len(tokens))
	for i, t := range tokens {
		actual[i] = t.Value
	}
	expected := []string{"node0", ";", "node1", ";", "block", "{", ";", "}", ";", "node2", "arg", ";", ""}
	require.Equal(t, expected, actual)
}
