package mast_test

import (
	. "github.com/fatlotus/mast"
	"strings"
	"testing"
)

var succeed = []struct {
	Parser Parser
	Source string
	Rep    string
}{
	{PEMDAS, "r=a+b*c", "r = (a + (b * c))"},
	{PEMDAS, "r=a*b+c", "r = ((a * b) + c)"},
	{PEMDAS, "r=a+b+c+d", "r = (((a + b) + c) + d)"},
	{PEMDAS, "x=a+b-c+d", "x = (((a + b) - c) + d)"},
	{PEMDAS, "x=-b", "x = (- b)"},
	{PEMDAS, "x=a+-b", "x = (a + (- b))"},
	{PEMDAS, "x=a''", "x = (' (' a))"},
	{PEMDAS, "r=a++++b", "r = (a + (+ (+ (+ b))))"},
	{PEMDAS, "r=a^b^c^d", "r = (a ^ (b ^ (c ^ d)))"},
	{PEMDAS, "r=Ax-b", "r = ((A x) - b)"},
	{PEMDAS, "r=xA-b", "r = (xA - b)"},
	{PEMDAS, "r=ABC", "r = ((A B) C)"},
	{PEMDAS, "err=trans * next - initial", "err = ((trans * next) - initial)"},
	{PEMDAS, "r=[([a+b])*c]", "r = ((a + b) * c)"},
	{PEMDAS, "x=inv(A) * b", "x = ((inv A) * b)"},
	{PEMDAS, "x = sin theta + z", "x = ((sin theta) + z)"},
	{PEMDAS, "x = sin A^H", "x = ((sin A) ^ H)"},
	{PEMDAS, "x = log a b c", "x = (((log a) b) c)"},
	{PEMDAS, "q, r = qr c", "(q , r) = (qr c)"},
	{PEMDAS, "x = 42", "x = 42"},
	{PEMDAS, "x = {}", "x = {}"},
	{PEMDAS, "x = {a}", "x = ({} a)"},
	{PEMDAS, "x = {a, b, c}", "x = ({} ((a , b) , c))"},
}

func TestParse(t *testing.T) {
	for _, test := range succeed {
		tree, err := test.Parser.Parse(test.Source)
		if err != nil {
			t.Errorf("%s, while parsing %#v", err, test.Source)
			continue
		}
		if tree.String() != test.Rep {
			t.Errorf("parsing %s\ngot       %#v;\nexpecting %#v",
				test.Source, tree.String(), test.Rep)
			continue
		}
	}
}

func TestParseExpr(t *testing.T) {
	for _, test := range succeed {
		source := strings.SplitN(test.Source, "=", 2)[1]
		tree, err := test.Parser.ParseExpr(source)
		if err != nil {
			t.Errorf("%s, while parsing %#v", err, source)
			continue
		}
		rep := strings.TrimSpace(strings.SplitN(test.Rep, "=", 2)[1])
		if tree.String() != rep {
			t.Errorf("parsing %s\ngot       %#v;\nexpecting %#v",
				source, tree.String(), rep)
			continue
		}
	}
}
