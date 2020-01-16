.PHONY: test
test: parser.go lexer.go
	go test -v .

parser.go: parser.y
	go tool yacc -v "" -o parser.go parser.y

lexer.go: lexer.nex
	nex -e -o lexer.go lexer.nex
	sed -i -e "s/.*NEX_END_OF_LEXER_STRUCT.*/  rules *map[string]rule/" lexer.go
	
