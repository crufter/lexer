package lexer

import (
	"fmt"
	"regexp"
	"strings"
)

type TokenExpr struct {
	Regex string
	Typ   int
}

type Token struct {
	Text  string
	Typ   int
	Occ   int
	Start int
}

type Tokens struct {
	S   []Token
	Pos int
}

func (t *Tokens) Has() bool {
	return t.Pos < len(t.S)-1
}

func (t *Tokens) Prev() Token {
	return t.S[t.Pos-2]
}

func (t *Tokens) Get() Token {
	v := t.S[t.Pos]
	t.Pos++
	return v
}

func (t *Tokens) Next() Token {
	return t.S[t.Pos]
}

func toMap(i ...int) map[int]struct{} {
	m := map[int]struct{}{}
	for _, v := range i {
		m[v] = struct{}{}
	}
	return m
}

func (t *Tokens) Until(i ...int) (Tokens, int) {
	m := toMap(i...)
	for i, v := range t.S[t.Pos+1:] {
		if _, ok := m[v.Typ]; ok {
			return Tokens{t.S[t.Pos+1 : i-1], 0}, i - 1
		}
	}
	return Tokens{}, -1
}

func (t *Tokens) Ignore(i int) {
	t.Pos += i
}

func LineAndPos(src string, pos int) (int, int) {
	lines := strings.Count(src[:pos], "\n")
	p := pos - strings.LastIndex(src[:pos], "\n")
	return lines, p
}

// Typ = 0 ignore
// Typ = -int means don't push subsequent repeating occurences of a token, so a b b b c becomes a b c.
// Token.Occ will contain how many times did it occur subsequently.
// (This allows for example to do whitespace significant lexing.)
func Lex(src string, tokens []TokenExpr) ([]Token, error) {
	pos := 0
	toks := []Token{}
	cache := []*regexp.Regexp{}
	for pos < len(src) {
		rem_src := src[pos:]
		match := false
		l := 0
		for i, v := range tokens {
			pattern, tag := v.Regex, v.Typ
			if len(cache) < i+1 {
				reg, err := regexp.Compile("^" + pattern)
				if err != nil {
					panic(err)
				}
				cache = append(cache, reg)
			}
			regex := cache[i]
			found := regex.FindIndex([]byte(rem_src))
			if found != nil {
				match = true
				l = found[1] - found[0]
				if tag > 0 {
					tok := Token{rem_src[found[0]:found[1]], tag, 0, found[0]}
					toks = append(toks, tok)
				}
				if tag < 0 { // If tag is negative, then avoid repeating occurences and push only once.
					if len(toks) == 0 || toks[len(toks)-1].Typ != tag { // If the previous token is not the same type as the current, then push.
						tok := Token{rem_src[found[0]:found[1]], tag, 1, found[0]}
						toks = append(toks, tok)
					} else { // If its the same type, increase the occurence counter of the already pushed one.
						toks[len(toks)-1].Occ++
					}
				}
				break
			}
		}
		if !match {
			err_line, err_pos := LineAndPos(src, pos)
			return toks, fmt.Errorf("%d:%d: Illegal character: %s", err_line, err_pos, string(src[pos]))
		} else {
			pos += l
		}
	}
	return toks, nil
}
