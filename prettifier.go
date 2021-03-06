package gotestdox

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

const eof rune = 0

var DebugWriter io.Writer = os.Stderr

// Prettify takes a string representing the name of a Go test, and attempts to
// turn it into a readable sentence, by replacing camel-case transitions and
// underscores with spaces.
//
// The input is expected to be a valid Go test name, as encoded by 'go test
// -json'. For example, the input might be: 'TestFoo/has_well-formed_output'.
//
// Here, the parent test is 'TestFoo', and this data is about a subtest whose
// name is 'has well-formed output'. Go replaces spaces in subtest names with
// underscores, and unprintable characters with the equivalent Go literal:
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.8:src/testing/match.go;l=133;drc=refs%2Ftags%2Fgo1.17.8
//
// Prettify does its best to undo this, yielding (something close to) the
// original subtest name. For example: "Foo has well-formed output".
//
// Multiword function names
//
// Because Go function names are often in camel-case, there's an ambiguity in
// parsing a test name like this: 'TestHandleInputClosesInputAfterReading'.
//
// We can see that this is about a function named 'HandleInput', but Prettify
// has no way of knowing that. To give it a hint, we can put an underscore after
// the name of the function. This will be interpreted as marking the end of a
// multiword function name. Example: "TestHandleInput_ClosesInputAfterReading".
//
// Debugging
//
// If the GOTESTDOX_DEBUG environment variable is set, Prettify will output
// (copious) debug information to os.Stderr, elaborating on its decisions.
func Prettify(input string) string {
	p := &prettifier{
		input: []rune(strings.TrimPrefix(input, "Test")),
		words: []string{},
		debug: io.Discard,
	}
	if os.Getenv("GOTESTDOX_DEBUG") != "" {
		p.debug = DebugWriter
	}
	p.log("input:", input)
	for state := betweenWords; state != nil; {
		state = state(p)
	}
	p.log("result:", p.words, "\n")
	return strings.Join(p.words, " ")
}

// Heavily inspired by Rob Pike's talk on 'Lexical Scanning in Go':
// https://www.youtube.com/watch?v=HxaD_trXwRE
type prettifier struct {
	debug          io.Writer
	input          []rune
	start, pos     int
	words          []string
	inSubTest      bool
	seenUnderscore bool
}

func (p *prettifier) backup() {
	p.pos--
}

func (p *prettifier) skip() {
	p.start = p.pos
}

func (p *prettifier) prev() rune {
	return p.input[p.pos-1]
}

func (p *prettifier) next() rune {
	next := p.peek()
	p.pos++
	return next
}

func (p *prettifier) peek() rune {
	if p.pos >= len(p.input) {
		return eof
	}
	next := p.input[p.pos]
	return next
}

func (p *prettifier) inInitialism() bool {
	for _, r := range p.input[p.start:p.pos] {
		if unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func (p *prettifier) emit() {
	word := string(p.input[p.start:p.pos])
	switch {
	case len(p.words) == 0:
		word = strings.Title(word)
	case len(word) == 1:
		word = strings.ToLower(word)
	case !p.inInitialism():
		word = strings.ToLower(word)
	}
	p.log(fmt.Sprintf("emit %q", word))
	p.words = append(p.words, word)
	p.skip()
}

func (p *prettifier) multiWordFunction() {
	var fname string
	for _, w := range p.words {
		fname += strings.Title(w)
	}
	p.log("multiword function", fname)
	p.words = []string{fname}
	p.seenUnderscore = true
}

func (p *prettifier) log(args ...interface{}) {
	fmt.Fprintln(p.debug, args...)
}

func (p *prettifier) logState(stateName string) {
	next := "EOF"
	if p.pos < len(p.input) {
		next = string(p.input[p.pos])
	}
	p.log(fmt.Sprintf("%s: [%s] -> %s",
		stateName,
		string(p.input[p.start:p.pos]),
		next,
	))
}

type stateFunc func(p *prettifier) stateFunc

func betweenWords(p *prettifier) stateFunc {
	for {
		p.logState("betweenWords")
		switch p.next() {
		case eof:
			return nil
		case '_', '/':
			p.skip()
		default:
			return inWord
		}
	}
}

func inWord(p *prettifier) stateFunc {
	for {
		p.logState("inWord")
		switch r := p.peek(); {
		case r == eof:
			p.emit()
			return nil
		case r == '_':
			p.emit()
			if !p.seenUnderscore && !p.inSubTest {
				// special 'end of function name' marker
				p.multiWordFunction()
			}
			return betweenWords
		case r == '/':
			p.emit()
			p.inSubTest = true
			return betweenWords
		case unicode.IsUpper(r):
			if p.prev() == '-' {
				// inside hyphenated word
				p.next()
				continue
			}
			if p.inInitialism() {
				// keep going
				p.next()
				continue
			}
			p.emit()
			return betweenWords
		case unicode.IsDigit(r):
			if unicode.IsDigit(p.prev()) {
				// in a multi-digit number
				p.next()
				continue
			}
			if p.prev() == '-' {
				// in a negative number
				p.next()
				continue
			}
			if p.prev() == '=' {
				// in some phrase like 'n=3'
				p.next()
				continue
			}
			if p.inInitialism() {
				// keep going
				p.next()
				continue
			}
			p.emit()
		default:
			if p.pos-p.start <= 1 {
				// word too short
				p.next()
				continue
			}
			if p.input[p.start] == '\'' {
				// inside a quoted word
				p.next()
				continue
			}
			if !p.inInitialism() {
				// keep going
				p.next()
				continue
			}
			// start a new word
			p.backup()
			p.emit()
		}
	}
}
