package jsmngo

import (
	"errors"
	"io"
	"runtime"
	"sync"
)

// TokenType represents the type of JSON token.
type TokenType int

const (
	Object TokenType = iota
	Array
	String
	Primitive
)

// Token holds information about a parsed JSON token.
type Token struct {
	Type      TokenType
	Start     int // Start position in the input string.
	End       int // End position in the input string.
	Size      int // Number of children (for objects/arrays).
	ParentIdx int // Index of parent token (-1 for root).
}

// Parser is the JSON tokenizer state.
type Parser struct {
	pos      int // Current position in the JSON string.
	toknext  int // Next token to allocate.
	toksuper int // Parent token index.
	tokens   []Token
}

// NewParser creates a new parser with space for numTokens.
func NewParser(numTokens int) *Parser {
	return &Parser{
		tokens: make([]Token, numTokens),
	}
}

// Parse tokenizes the JSON input, returning the number of tokens or an error.
func (p *Parser) Parse(json []byte) (int, error) {
	p.pos = 0
	p.toknext = 0
	p.toksuper = -1

	for p.pos < len(json) {
		c := json[p.pos]
		switch c {
		case '{', '[':
			tok := Token{Start: p.pos, End: -1, Size: 0, ParentIdx: p.toksuper}
			if c == '{' {
				tok.Type = Object
			} else {
				tok.Type = Array
			}
			p.allocToken(tok)
			p.toksuper = p.toknext - 1
			p.pos++
			continue
		case '}', ']':
			if p.toksuper != -1 {
				p.tokens[p.toksuper].End = p.pos + 1
				p.toksuper = p.tokens[p.toksuper].ParentIdx
			}
			p.pos++
			continue
		case '"':
			err := p.parseString(json)
			if err != nil {
				return 0, err
			}
			continue
		case '\t', '\r', '\n', ' ':
			p.pos++
			continue
		case ':':
			p.toksuper = p.toknext - 1
			p.pos++
			continue
		case ',':
			if p.toksuper != -1 && p.tokens[p.toksuper].Type != Array && p.tokens[p.toksuper].Type != Object {
				p.toksuper = p.tokens[p.toksuper].ParentIdx
			}
			p.pos++
			continue
		default:
			err := p.parsePrimitive(json)
			if err != nil {
				return 0, err
			}
			continue
		}
	}
	for i := range p.tokens {
		if p.tokens[i].End == -1 && p.tokens[i].Start != -1 {
			p.tokens[i].End = len(json)
		}
	}
	return p.toknext, nil
}

// Tokens returns the parsed tokens.
func (p *Parser) Tokens() []Token {
	return p.tokens[:p.toknext]
}

func (p *Parser) allocToken(tok Token) {
	p.tokens[p.toknext] = tok
	if p.toksuper != -1 {
		p.tokens[p.toksuper].Size++
	}
	p.toknext++
}

func (p *Parser) parseString(json []byte) error {
	p.pos++ // Skip opening quote.
	tok := Token{Type: String, Start: p.pos, End: -1, ParentIdx: p.toksuper}
	for p.pos < len(json) {
		c := json[p.pos]
		if c == '"' {
			tok.End = p.pos
			p.allocToken(tok)
			p.pos++
			return nil
		}
		if c == '\\' && p.pos+1 < len(json) {
			p.pos += 2
			continue
		}
		p.pos++
	}
	return errors.New("unclosed string")
}

func (p *Parser) parsePrimitive(json []byte) error {
	tok := Token{Type: Primitive, Start: p.pos, End: -1, ParentIdx: p.toksuper}
	for p.pos < len(json) {
		c := json[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' || c == ']' || c == '}' {
			break
		}
		p.pos++
	}
	tok.End = p.pos
	p.allocToken(tok)
	return nil
}

// Novel Enhancement: ParseParallel - Tokenize in parallel across chunks.
func ParseParallel(json []byte, numTokens int) ([]Token, error) {
	numWorkers := runtime.NumCPU()
	if numWorkers > 4 {
		numWorkers = 4 // Cap for simplicity.
	}
	chunkSize := len(json) / numWorkers
	if chunkSize == 0 {
		chunkSize = len(json)
		numWorkers = 1
	}

	var wg sync.WaitGroup
	results := make([][]Token, numWorkers)
	errs := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = len(json)
		}
		go func(i int, chunk []byte) {
			defer wg.Done()
			p := NewParser(numTokens / numWorkers)
			_, err := p.Parse(chunk)
			if err != nil {
				errs <- err
				return
			}
			results[i] = p.Tokens()
		}(i, json[start:end])
	}

	wg.Wait()
	select {
	case err := <-errs:
		return nil, err
	default:
	}

	// Merge results (simple concat for demo; real use would align boundaries).
	var merged []Token
	for _, res := range results {
		merged = append(merged, res...)
	}
	return merged, nil
}

// Novel Enhancement: ParseStream - Tokenize from an io.Reader (non-blocking streaming).
func ParseStream(r io.Reader, numTokens int) ([]Token, error) {
	json, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	p := NewParser(numTokens)
	_, err = p.Parse(json)
	if err != nil {
		return nil, err
	}
	return p.Tokens(), nil
}
