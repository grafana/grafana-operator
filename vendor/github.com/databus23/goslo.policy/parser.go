package policy

import __yyfmt__ "fmt"

//line parser.y:2
import (
	"fmt"
)

func (y yySymType) String() string {
	return fmt.Sprintf("{str:%s,...}", y.str)
}

//line parser.y:14
type yySymType struct {
	yys int
	f   rule
	str string
	num int
	b   bool
}

const and = 57346
const or = 57347
const not = 57348
const variable = 57349
const unquotedStr = 57350
const constStr = 57351
const number = 57352
const boolean = 57353

var yyToknames = []string{
	"'('",
	"')'",
	"'@'",
	"'!'",
	"':'",
	"and",
	"or",
	"not",
	"variable",
	"unquotedStr",
	"constStr",
	"number",
	"boolean",
}
var yyStatenames = []string{}

const yyEOFCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line parser.y:112

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 14
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 25

var yyAct = []int{

	4, 22, 8, 9, 21, 19, 20, 3, 2, 6,
	7, 18, 12, 13, 11, 11, 10, 11, 10, 16,
	17, 15, 14, 1, 5,
}
var yyPact = []int{

	-4, -1000, 8, -4, -4, -1000, 14, 13, -1000, -1000,
	-4, -4, -1000, 6, -8, -11, 5, -1000, -1000, -1000,
	-1000, -1000, -1000,
}
var yyPgo = []int{

	0, 24, 8, 23,
}
var yyR1 = []int{

	0, 3, 3, 2, 2, 2, 2, 2, 1, 1,
	1, 1, 1, 1,
}
var yyR2 = []int{

	0, 0, 1, 2, 3, 3, 3, 1, 3, 3,
	3, 3, 1, 1,
}
var yyChk = []int{

	-1000, -3, -2, 11, 4, -1, 13, 14, 6, 7,
	10, 9, -2, -2, 8, 8, -2, -2, 5, 13,
	14, 12, 12,
}
var yyDef = []int{

	1, -2, 2, 0, 0, 7, 0, 0, 12, 13,
	0, 0, 3, 0, 0, 0, 5, 6, 4, 8,
	9, 10, 11,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 7, 3, 3, 3, 3, 3, 3,
	4, 5, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 8, 3,
	3, 3, 3, 3, 6,
}
var yyTok2 = []int{

	2, 3, 9, 10, 11, 12, 13, 14, 15, 16,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(c), uint(char))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEOFCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		//line parser.y:38
		{
			var f rule = func(c Context) bool { return true }
			yylex.(*lexer).parseResult = f
		}
	case 2:
		//line parser.y:44
		{
			yylex.(*lexer).parseResult = yyS[yypt-0].f
		}
	case 3:
		//line parser.y:50
		{
			f := yyS[yypt-0].f
			yyVAL.f = func(c Context) bool { return !f(c) }
		}
	case 4:
		//line parser.y:56
		{
			f := yyS[yypt-1].f
			yyVAL.f = func(c Context) bool { return f(c) }
		}
	case 5:
		//line parser.y:62
		{
			left, right := yyS[yypt-2].f, yyS[yypt-0].f
			yyVAL.f = func(c Context) bool { return left(c) || right(c) }
		}
	case 6:
		//line parser.y:68
		{
			left, right := yyS[yypt-2].f, yyS[yypt-0].f
			yyVAL.f = func(c Context) bool { return left(c) && right(c) }
		}
	case 7:
		//line parser.y:74
		{
			yyVAL.f = yyS[yypt-0].f
		}
	case 8:
		//line parser.y:80
		{
			left, right := yyS[yypt-2].str, yyS[yypt-0].str
			yyVAL.f = func(c Context) bool { return c.genericCheck(left, right, false) }
		}
	case 9:
		//line parser.y:86
		{
			left, right := yyS[yypt-2].str, yyS[yypt-0].str
			yyVAL.f = func(c Context) bool { return c.genericCheck(left, right, false) }
		}
	case 10:
		//line parser.y:92
		{
			left, right := yyS[yypt-2].str, yyS[yypt-0].str
			yyVAL.f = func(c Context) bool { return c.genericCheck(left, right, true) }
		}
	case 11:
		//line parser.y:98
		{
			left, right := yyS[yypt-2].str, yyS[yypt-0].str
			yyVAL.f = func(c Context) bool { return c.checkVariable(right, left) }
		}
	case 12:
		//line parser.y:104
		{
			yyVAL.f = func(_ Context) bool { return true }
		}
	case 13:
		//line parser.y:109
		{
			yyVAL.f = func(_ Context) bool { return false }
		}
	}
	goto yystack /* stack new state and value */
}
