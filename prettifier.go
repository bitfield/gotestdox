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

func (p *prettifier) next() rune {
	if p.pos >= len(p.input) {
		return eof
	}
	next := p.input[p.pos]
	p.pos++
	return next
}

func (p *prettifier) emit(c wordCase) {
	word := string(p.input[p.start:p.pos])
	if word == "" {
		return
	}
	switch {
	case len(p.words) == 0:
		word = strings.Title(word)
	case c == allLower:
		word = strings.ToLower(word)
	case c == allUpper:
		word = strings.ToUpper(word)
	}
	p.log(fmt.Sprintf("emit %q", word))
	p.words = append(p.words, word)
	p.start = p.pos
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
		case r == eof:
			return nil
		case unicode.IsUpper(r):
			return inWordUpper
		// case r == '/':
		// 	p.inSubTest = true
		// 	return betweenWords
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
		case unicode.IsUpper(r):
			return inWordUpper
		case unicode.IsLower(r):
			return inWordLower
			// case unicode.IsLower(r):
			// 	p.emitRune(r)
			// 	return inWordLower
			// case r == '_':
			// 	p.seenUnderscore = true
			// 	return betweenWords
			// case r == '/':
			// 	p.inSubTest = true
			// 	return betweenWords
			// default:
			// 	p.emitRune(r)
			// 	return inWordLower
		}
	}
}

func inInitialism(p *prettifier) stateFunc {
	for {
		p.debugState("inInitialism")
		switch r := p.next(); {
		case r == eof:
			p.emit(allUpper)
			return nil
		case r == '_':
			p.backup()
			p.emit(allUpper)
			if !p.seenUnderscore && !p.inSubTest {
				p.multiWordFunction()
				return betweenWords
			}
			p.skip()
			return betweenWords
		case unicode.IsLower(r):
			p.backup()
			p.backup()
			wordCase := allUpper
			if (p.pos - p.start) == 1 {
				wordCase = allLower // don't capitalise single-letter words
			}
			p.emit(wordCase)
			return inWordLower
			// case unicode.IsLower(r):
			// 	// If we see a lowercase rune, it means we're already one rune
			// 	// into the next word. We need to emit the previous word, and
			// 	// reset the current word to be just the previous rune, before
			// 	// adding the current rune to it.
			// 	prev := p.curWord[:len(p.curWord)-1]
			// 	if len(prev) == 1 {
			// 		prev = strings.ToLower(prev)
			// 	}
			// 	p.log("emit", prev)
			// 	p.words = append(p.words, prev)
			// 	p.curWord = strings.ToLower(string(p.curWord[len(p.curWord)-1]))
			// 	p.emitRune(r)
			// 	return inWordLower
			// case unicode.IsUpper(r):
			// 	// A hyphen marks the end of an initialism
			// 	if strings.Contains(p.curWord, "-") {
			// 		p.emitRune(unicode.ToLower(r))
			// 		return inWordLower
			// 	}
			// 	p.emitRune(r)
			// 	return inInitialism
			// case r == '_':
			// 	p.emitWord()
			// 	// If this is the first underscore we've seen, and we're not yet
			// 	// in a subtest name, then treat all the words we've seen so far
			// 	// as making up the name of a multiword function ('HandleInput')
			// 	if !p.seenUnderscore && !p.inSubTest {
			// 		p.multiWordFunction()
			// 		p.seenUnderscore = true
			// 	}
			// 	return inWordLower
			// case r == '/':
			// 	// We're entering a subtest name, so from now on the multiword
			// 	// function rule will be disabled
			// 	p.inSubTest = true
			// 	p.emitWord()
			// 	return betweenWords
			// default:
			// 	p.emitRune(r)
			// 	return inInitialism
			// }
		}
	}
}

func inWordUpper(p *prettifier) stateFunc {
	p.debugState("inWordUpper")
	switch r := p.next(); {
	case r == eof:
		p.emit(allLower)
		return nil
	case unicode.IsUpper(r):
		return inInitialism
	// case unicode.IsUpper(r), unicode.IsDigit(r):
	// 	p.pos++
	// 	return inInitialism
	// case r == '_':
	// 	p.emitWord()
	// 	if !p.seenUnderscore && !p.inSubTest {
	// 		p.multiWordFunction()
	// 		p.seenUnderscore = true
	// 	}
	// 	return inWordLower
	// case r == '/':
	// 	p.inSubTest = true
	// 	p.emitWord()
	// 	return betweenWords
	default:
		return inWordLower
		// 	p.curWord = strings.ToLower(p.curWord)
		// 	p.emitRune(r)
		// 	return inWordLower
	}
}

func inWordLower(p *prettifier) stateFunc {
	for {
		p.debugState("inWordLower")
		switch r := p.next(); {
		case r == eof:
			p.emit(allLower)
			return nil
		case r == '_':
			p.backup()
			p.emit(allLower)
			if !p.seenUnderscore && !p.inSubTest {
				p.multiWordFunction()
				return betweenWords
			}
			p.skip()
			return betweenWords
		case r == '/':
			p.backup()
			p.emit(allLower)
			p.inSubTest = true
			return betweenWords
		case unicode.IsUpper(r):
			p.backup()
			p.emit(allLower)
			return betweenWords
		case unicode.IsDigit(r):
			p.backup()
			p.emit(allLower)
			return inNumber
			// case unicode.IsUpper(r):
			// 	if strings.HasSuffix(p.curWord, "-") {
			// 		p.emitRune(r)
			// 		return inWordUpper
			// 	}
			// 	p.emitWord()
			// 	p.emitRune(r)
			// 	return inWordUpper
			// case unicode.IsDigit(r):
			// 	if !strings.HasSuffix(p.curWord, "-") && !strings.HasSuffix(p.curWord, "=") {
			// 		p.emitWord()
			// 	}
			// 	p.emitRune(r)
			// 	return inNumber
			// case r == '_':
			// 	p.emitWord()
			// 	if !p.seenUnderscore && !p.inSubTest {
			// 		p.multiWordFunction()
			// 		p.seenUnderscore = true
			// 	}
			// 	return inWordLower
			// case r == '/':
			// 	p.inSubTest = true
			// 	p.emitWord()
			// 	return betweenWords
			// default:
			// 	p.emitRune(r)
			// 	return inWordLower
		}
	}
}

func inNumber(p *prettifier) stateFunc {
	for {
		p.debugState("inNumber")
		switch r := p.next(); {
		case r == eof:
			return nil
		case r == '_':
			p.backup()
			p.emit(allLower)
			return betweenWords
		case unicode.IsUpper(r):
			p.backup()
			p.emit(allLower)
			return inWordUpper
			// 	case unicode.IsDigit(r):
			// 		p.emitRune(r)
			// 		return inNumber
			// 	case r == '_':
			// 		p.emitWord()
			// 		return betweenWords
			// 	case r == '/':
			// 		p.inSubTest = true
			// 		p.emitWord()
			// 		return betweenWords
			// 	default:
			// 		p.emitRune(r)
			// 		return betweenWords
		}
	}
}
