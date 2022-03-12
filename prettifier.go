package gotestdox

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

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
func Prettify(tname string) string {
	tname = strings.TrimPrefix(tname, "Test")
	p := &prettifier{
		input: []rune(tname),
		words: []string{},
	}
	if os.Getenv("GOTESTDOX_DEBUG") != "" {
		p.debug = os.Stderr
	}
	for state := start; state != nil; {
		state = state(p)
	}
	p.log(fmt.Sprintf("%#v\n", p.words))
	return strings.Join(p.words, " ")
}

// This lexer implementation owes a lot, if not everything, to Rob Pike's talk
// on 'Lexical Scanning in Go': https://www.youtube.com/watch?v=HxaD_trXwRE
type prettifier struct {
	debug          io.Writer
	input          []rune
	start, pos     int
	words          []string
	inSubTest      bool
	seenUnderscore bool
}

const eof rune = 0

type wordCase int

const (
	allLower wordCase = iota
	allUpper
)

func (p *prettifier) backup() {
	p.pos--
}

func (p *prettifier) skip() {
	p.start = p.pos
}

func (p *prettifier) prev() rune {
	if p.pos == 0 {
		return p.input[0]
	}
	return p.input[p.pos-1]
}

func (p *prettifier) next() rune {
	if p.pos >= len(p.input) {
		return eof
	}
	next := p.input[p.pos]
	p.pos++
	return next
}

func (p *prettifier) inHyphenation() bool {
	for _, r := range p.input[p.start:p.pos] {
		if r == '-' {
			return true
		}
	}
	return false
}

func (p *prettifier) inInitialism() bool {
	for _, r := range p.input[p.start:p.pos] {
		if unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func (p *prettifier) emit(c wordCase) {
	word := string(p.input[p.start:p.pos])
	if word == "" {
		return
	}
	switch {
	case len(p.words) == 0:
		word = strings.Title(word)
	case len(word) == 1:
		word = strings.ToLower(word)
	case c == allLower:
		word = strings.ToLower(word)
	case c == allUpper:
		word = strings.ToUpper(word)
	}
	p.log(fmt.Sprintf("emit %q", word))
	p.words = append(p.words, word)
	p.start = p.pos
}

func (p *prettifier) emitUpper() {
	p.emit(allUpper)
}

func (p *prettifier) emitLower() {
	p.emit(allLower)
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
	if p.debug == nil {
		return
	}
	fmt.Fprintln(p.debug, args...)
}

func (p *prettifier) debugState(stateName string) {
	if p.debug == nil {
		return
	}
	next := "EOF"
	if p.pos < len(p.input) {
		next = string(p.input[p.pos])
	}
	fmt.Fprintf(p.debug, "%s: [%s] -> %s\n",
		stateName,
		string(p.input[p.start:p.pos]),
		next,
	)
}

type stateFunc func(p *prettifier) stateFunc

func start(p *prettifier) stateFunc {
	for {
		p.debugState("start")
		switch r := p.next(); {
		case r == '_':
			p.skip()
		case r == '/':
			p.skip()
			p.inSubTest = true
			return betweenWords
		case r == eof:
			return nil
		case unicode.IsUpper(r):
			return inWord
		default:
			p.pos++
			return start
		}
	}
}

func betweenWords(p *prettifier) stateFunc {
	for {
		p.debugState("betweenWords")
		switch r := p.next(); {
		case r == '_':
			p.skip()
		case r == '/':
			p.skip()
			p.inSubTest = true
		case r == '-':
			return inNumber
		case unicode.IsLetter(r):
			return inWord
		}
	}
}

func inWord(p *prettifier) stateFunc {
	for {
		p.debugState("inWord")
		switch r := p.next(); {
		case r == eof:
			if p.inInitialism() {
				p.emitUpper()
			} else {
				p.emitLower()
			}
			return nil
		case r == '_':
			p.backup()
			if p.inInitialism() {
				p.emitUpper()
			} else {
				p.emitLower()
			}
			if !p.seenUnderscore && !p.inSubTest {
				p.multiWordFunction()
				return betweenWords
			}
			p.skip()
			return betweenWords
		case r == '/':
			p.backup()
			p.emitLower()
			p.inSubTest = true
			return betweenWords
		case unicode.IsLower(r):
			p.backup()
			if p.inInitialism() && (p.pos-p.start) > 1 {
				p.backup()
				p.emitUpper()
				return inWord
			}
			p.next()
		case unicode.IsUpper(r):
			p.backup()
			if p.prev() != '-' && !p.inInitialism() {
				p.emitLower()
				return betweenWords
			}
			p.next()
		case unicode.IsDigit(r):
			p.backup()
			if p.prev() != '-' && p.prev() != '=' && !p.inInitialism() {
				p.emitLower()
			}
			return inNumber
		}
	}
}

func inNumber(p *prettifier) stateFunc {
	for {
		p.debugState("inNumber")
		switch r := p.next(); {
		case r == eof:
			p.emitLower()
			return nil
		case r == '_':
			p.backup()
			p.emitLower()
			return betweenWords
		case unicode.IsUpper(r):
			p.backup()
			if !p.inInitialism() {
				p.emitLower()
			}
			return inWord
		}
	}
}
