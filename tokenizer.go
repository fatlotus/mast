package mast

import (
	"unicode"
)

func isWsp(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v'
}

type runePred func(rune) bool

var tokenPreds = []runePred{
	unicode.IsLetter,
	isWsp,
}

func (p Parser) tokenize(code string) ([]string, error) {
	st := runePred(nil)
	matches := []string{}
	buf := []rune{}

	for _, c := range code {
		if st != nil && st(c) {
			buf = append(buf, c)
		} else {
			if len(buf) > 0 && !isWsp(buf[0]) {
				matches = append(matches, string(buf))
			}

			st = nil
			for _, pred := range tokenPreds {
				if pred(c) {
					st = pred
					break
				}
			}
			buf = buf[:0]
			buf = append(buf, c)
		}
	}

	matches = append(matches, string(buf))
	matches = append(matches, "") // eof marker
	return matches, nil
}
