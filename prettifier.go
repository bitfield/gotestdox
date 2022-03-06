package gotestdox

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type prettifier struct {
	debug          io.Writer
	curWord        string
	words          []string
	inSubTest      bool
	seenUnderscore bool
}

func (p *prettifier) emitRune(r rune) {
	p.curWord += string(r)
}

func (p *prettifier) emitWord() {
	if p.curWord == "" {
		return
	}
	p.log("emit", p.curWord)
	p.words = append(p.words, p.curWord)
	p.curWord = ""
}

func (p *prettifier) multiWordFunction() {
	var fname string
	for _, w := range p.words {
		fname += strings.Title(w)
	}
	p.log("multiword function", fname)
	p.words = []string{fname}
}

func (p *prettifier) log(args ...interface{}) {
	if p.debug != nil {
		fmt.Fprintln(p.debug, args...)
	}
}

type stateFunc func(p *prettifier, r rune) stateFunc

func start(p *prettifier, r rune) stateFunc {
	p.log("start", string(r))
	switch {
	case unicode.IsUpper(r):
		p.emitRune(r)
		return inFirstWord
	default:
		return start
	}
}

func betweenWords(p *prettifier, r rune) stateFunc {
	p.log("betweenWords", p.curWord, string(r))
	switch {
	case unicode.IsUpper(r):
		p.emitRune(r)
		return inWordUpper
	case unicode.IsLower(r):
		p.emitRune(r)
		return inWordLower
	default:
		p.emitRune(r)
		return inWordLower
	}
}

func inInitialism(p *prettifier, r rune) stateFunc {
	p.log("inInitialism", p.curWord, string(r))
	switch {
	case unicode.IsLower(r):
		prev := p.curWord[:len(p.curWord)-1]
		if len(prev) == 1 {
			prev = strings.ToLower(prev)
		}
		p.log("emit", prev)
		p.words = append(p.words, prev)
		p.curWord = strings.ToLower(string(p.curWord[len(p.curWord)-1]) + string(r))
		return inWordLower
	case unicode.IsUpper(r):
		if strings.Contains(p.curWord, "-") {
			p.emitRune(unicode.ToLower(r))
			return inWordLower
		}
		p.emitRune(r)
		return inInitialism
	case r == '_':
		p.emitWord()
		if !p.seenUnderscore && !p.inSubTest {
			p.multiWordFunction()
			p.seenUnderscore = true
		}
		return inWordLower
	default:
		p.emitRune(r)
		return inInitialism
	}
}

func inFirstWord(p *prettifier, r rune) stateFunc {
	p.log("inFirstWord", p.curWord, string(r))
	switch {
	case unicode.IsUpper(r):
		p.emitRune(r)
		return inInitialism
	default:
		p.emitRune(r)
		return inWordLower
	}
}

func inWordUpper(p *prettifier, r rune) stateFunc {
	p.log("inWordUpper", p.curWord, string(r))
	switch {
	case unicode.IsUpper(r):
		p.emitRune(r)
		return inInitialism
	default:
		p.curWord = strings.ToLower(p.curWord)
		p.emitRune(r)
		return inWordLower
	}
}

func inWordLower(p *prettifier, r rune) stateFunc {
	p.log("inWordLower", p.curWord, string(r))
	switch {
	case unicode.IsUpper(r):
		if strings.HasSuffix(p.curWord, "-") {
			p.emitRune(r)
			return inWordUpper
		}
		p.emitWord()
		p.emitRune(r)
		return inWordUpper
	case unicode.IsDigit(r):
		if !strings.HasSuffix(p.curWord, "-") {
			p.emitWord()
		}
		p.emitRune(r)
		return inNumber
	case r == '_':
		p.emitWord()
		if !p.seenUnderscore && !p.inSubTest {
			p.multiWordFunction()
			p.seenUnderscore = true
		}
		return inWordLower
	case r == '/':
		p.inSubTest = true
		p.emitWord()
		return betweenWords
	default:
		p.emitRune(r)
		return inWordLower
	}
}

func inNumber(p *prettifier, r rune) stateFunc {
	p.log("inNumber", p.curWord, string(r))
	switch {
	case unicode.IsDigit(r):
		p.emitRune(r)
		return inNumber
	case unicode.IsUpper(r):
		p.emitWord()
		p.emitRune(r)
		return inWordUpper
	case r == '_':
		p.emitWord()
		return betweenWords
	default:
		p.emitRune(r)
		return betweenWords
	}
}

func NewPrettifier() *prettifier {
	return &prettifier{
		words: []string{},
	}
}

func Prettify(tname string) string {
	tname = strings.TrimPrefix(tname, "Test")
	p := NewPrettifier()
	if os.Getenv("GOTESTDOX_DEBUG") != "" {
		p.debug = os.Stderr
	}
	state := start
	for _, r := range tname {
		state = state(p, r)
	}
	p.emitWord()
	p.log(fmt.Sprintf("%#v\n", p.words))
	return strings.Join(p.words, " ")
}
