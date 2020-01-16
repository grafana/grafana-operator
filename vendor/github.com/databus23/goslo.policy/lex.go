package policy

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

func (l *lexer) Error(e string) {
	l.parseResult = fmt.Sprintf("%s. Column: %d", e, l.Column())
}

func (l *lexer) Lex(lval *yySymType) int {
	i := <-l.items
	lval.str = i.val
	switch i.typ {
	case itemEOF:
		return 0
	case itemLeftParen:
		return '('
	case itemRightParen:
		return ')'
	case itemTrueCheck:
		return '@'
	case itemFalseCheck:
		return '!'
	case itemColon:
		return ':'
	case itemAnd:
		return and
	case itemOr:
		return or
	case itemNot:
		return not
	case itemVariable:
		return variable
	case itemString:
		return unquotedStr
	case itemConstString:
		return constStr
	case itemNumber:
		lval.num, _ = strconv.Atoi(i.val)
		return number
	case itemBool:
		lval.b = i.val == "True"
		return boolean

	}
	//unknown token
	return 1
}
func (l *lexer) Column() int {
	return l.start
}

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos int      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemVariable:
		return fmt.Sprintf("request[%s]", i.val)
	case i.typ == itemConstString:
		return fmt.Sprintf("`%s`", i.val)
	}
	for str, typ := range items {
		if typ == i.typ {
			return str
		}
	}
	return i.val
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemAnd
	itemOr
	itemNot
	itemTrueCheck  // @
	itemFalseCheck // !
	itemLeftParen
	itemRightParen
	itemString      // user_id, rule
	itemConstString // "blfasel", 'foo'
	itemVariable    // %(target.user_id)s
	itemColon
	itemBool // True, False
	itemNumber
)

var items = map[string]itemType{
	"and":   itemAnd,
	"not":   itemNot,
	"or":    itemOr,
	"!":     itemFalseCheck,
	"@":     itemTrueCheck,
	"True":  itemConstString,
	"False": itemConstString,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input string    // the string being scanned
	state stateFn   // the next lexing function to enter
	pos   int       // current position in the input
	start int       // start position of this item
	width int       // width of last rune read from input
	items chan item // channel of scanned items

	parseResult interface{}
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexRule; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func lexRule(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemEOF)
			return nil
		case isSpace(r):
			//skip spaces
			l.ignore()
		case r == '%' && l.peek() == '(':
			l.next()
			return lexVariable
		case r == '(':
			l.emit(itemLeftParen)
		case r == ')':
			l.emit(itemRightParen)
		case r == ':':
			l.emit(itemColon)
		case r == '"' || r == '\'':
			return lexQuotedString
		default:
			l.backup()
			return lexItem
		}
	}
}

// can be @ !, and, or or an unqouted string
func lexItem(l *lexer) stateFn {
	for {
		//stop if we reached a terminator
		if r := l.next(); isSpace(r) || r == eof || r == ':' || r == ')' || r == '(' {
			l.backup()
			break
		}
	}
	word := l.input[l.start:l.pos]
	if i, ok := items[word]; ok {
		l.emit(i)
	} else {
		if re := regexp.MustCompile(`^[-+]?\d+$`); re.MatchString(word) {
			l.emit(itemConstString)
		} else {
			l.emit(itemString)
		}
	}
	return lexRule
}

func lexVariable(l *lexer) stateFn {
	//discard already lexed %(
	l.ignore()
	for {
		r := l.next()
		if isSpace(r) || r == eof {
			return l.errorf("unterminated variable")
		}
		if r == ')' && l.peek() == 's' {
			l.backup()
			l.emit(itemVariable)
			l.next()
			l.next()
			l.ignore()
			return lexRule
		}
	}
}

func lexQuotedString(l *lexer) stateFn {
	//delimiter can only by " or '
	delimiter, _ := utf8.DecodeRuneInString(l.input[l.start:])
	l.ignore()
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated character constant")
		case delimiter:
			break Loop
		}
	}
	l.backup()
	l.emit(itemConstString)
	l.next()
	l.ignore()
	return lexRule
}

func newLexer(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
