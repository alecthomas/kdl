package kdl

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"
)

var (
	lex = lexer.MustStateful(lexer.Rules{
		`Root`: {
			{`RawStringStart`, `r(#*)"`, lexer.Push(`RawString`)},
			{`Ident`, `[-_!\p{L}][-_\p{L}0-9]*`, nil},
			{`String`, `"(\\.|[^"])*"`, nil},
			{`Number`, `\d+`, nil},
			{`Punct`, `[\\{};=]`, nil},
			{`Skip`, `/-`, nil},
			{`NL`, `\n\r?`, nil},
			{`startComment`, `/\*`, lexer.Push(`Comment`)},
			{`singleLineComment`, `//.*`, nil},
			{`whitespace`, `[ \t]+`, nil},
		},
		`RawString`: {
			{`RawStringEnd`, `"\1`, lexer.Pop()},
			{`RawStringText`, `[^"]*`, nil},
		},
		`Comment`: {
			{`startInnerComment`, `/\*`, lexer.Push(`Comment`)},
			{`endComment`, `\*/`, lexer.Pop()},
			{`commentText`, `[^/*]+`, nil},
		},
	})
	identToken     = lex.Symbols()["Ident"]
	stringEndToken = lex.Symbols()["RawStringEnd"]
	stringToken    = lex.Symbols()["String"]
	numberToken    = lex.Symbols()["Number"]
)

// A Lexer that inserts semi-colons and collapses \-separated lines.
type fixupLexerDefinition struct{}

func (l fixupLexerDefinition) Lex(path string, r io.Reader) (lexer.Lexer, error) { // nolint: golint
	ll, err := lex.Lex(path, r)
	if err != nil {
		return nil, err
	}
	return &fixupLexer{lexer: ll}, nil
}

func (l fixupLexerDefinition) Symbols() map[string]lexer.TokenType { // nolint: golint
	return lex.Symbols()
}

type fixupLexer struct {
	lexer lexer.Lexer
	last  lexer.Token
	next  *lexer.Token
	eof   bool
}

func (l *fixupLexer) Next() (lexer.Token, error) {
next:
	for {
		if l.eof {
			return lexer.EOFToken(l.last.Pos), nil
		}
		var token lexer.Token
		if l.next != nil {
			token = *l.next
			l.next = nil
		} else {
			var err error
			token, err = l.lexer.Next()
			if err != nil {
				return token, err
			}
			// Always insert a ; before }
			if token.Value == "}" {
				l.next = &token
				return lexer.Token{Type: ';', Value: ";", Pos: token.Pos}, nil
			}
		}

		l.eof = token.EOF()

		// Delete \\ followed by \n
		if token.Value == "\\" {
			l.last = token
			continue next
		}
		if token.Value != "\n" && !token.EOF() {
			l.last = token
			return token, nil
		}

		// Do we need to insert a semi-colon?
		if l.last.Value == "}" {
			token.Value = ";"
			token.Type = ';'
		} else {
			switch l.last.Type {
			case numberToken, stringEndToken, identToken, stringToken:
				token.Value = ";"
				token.Type = ';'

			default:
				l.last = token
				continue next
			}
		}
		l.last = token
		return token, nil
	}
}
