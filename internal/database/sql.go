package database

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenKind uint8

const (
	tokenIdentifier tokenKind = iota
	tokenNumber
	tokenString
	tokenSymbol
	tokenEOF
)

type token struct {
	kind  tokenKind
	value string
	pos   int
}

type literalKind uint8

const (
	literalString literalKind = iota
	literalNumber
	literalBoolean
	literalNull
)

type sqlLiteral struct {
	kind  literalKind
	value string
}

type statement interface {
	isStatement()
}

type createTableStatement struct {
	table   string
	columns []Column
}

func (createTableStatement) isStatement() {}

type dropTableStatement struct{ table string }

func (dropTableStatement) isStatement() {}

type insertStatement struct {
	table   string
	columns []string
	values  []sqlLiteral
}

func (insertStatement) isStatement() {}

// predicateOp is the comparison a WHERE clause applies. v0.1 supports one
// predicate per statement.
type predicateOp uint8

const (
	predicateEqual predicateOp = iota
	predicateIsNull
	predicateIsNotNull
)

type predicate struct {
	column string
	op     predicateOp
	value  sqlLiteral
}

type selectStatement struct {
	table      string
	columns    []string
	allColumns bool
	where      *predicate
}

func (selectStatement) isStatement() {}

type assignment struct {
	column string
	value  sqlLiteral
}

type updateStatement struct {
	table       string
	assignments []assignment
	where       *predicate
}

func (updateStatement) isStatement() {}

type deleteStatement struct {
	table string
	where *predicate
}

func (deleteStatement) isStatement() {}

func parseSQL(input string) (statement, error) {
	tokens, err := lex(input)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	stmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	p.acceptSymbol(";")
	if p.peek().kind != tokenEOF {
		return nil, p.syntaxError(p.peek(), "unexpected token %q", p.peek().value)
	}
	return stmt, nil
}

func lex(input string) ([]token, error) {
	tokens := make([]token, 0, 16)
	runes := []rune(input)
	for i := 0; i < len(runes); {
		if unicode.IsSpace(runes[i]) {
			i++
			continue
		}

		start := i
		switch {
		case unicode.IsLetter(runes[i]) || runes[i] == '_':
			i++
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			tokens = append(tokens, token{kind: tokenIdentifier, value: string(runes[start:i]), pos: start})
		case unicode.IsDigit(runes[i]) || (runes[i] == '-' && i+1 < len(runes) && unicode.IsDigit(runes[i+1])):
			i++
			dotSeen := false
			for i < len(runes) {
				if unicode.IsDigit(runes[i]) {
					i++
					continue
				}
				if runes[i] == '.' && !dotSeen {
					dotSeen = true
					i++
					continue
				}
				break
			}
			i = consumeExponent(runes, i)
			tokens = append(tokens, token{kind: tokenNumber, value: string(runes[start:i]), pos: start})
		case runes[i] == '\'':
			i++
			var value strings.Builder
			closed := false
			for i < len(runes) {
				if runes[i] == '\'' {
					if i+1 < len(runes) && runes[i+1] == '\'' {
						value.WriteRune('\'')
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				value.WriteRune(runes[i])
				i++
			}
			if !closed {
				return nil, newError(CodeSyntax, "unterminated string at position %d", start+1)
			}
			tokens = append(tokens, token{kind: tokenString, value: value.String(), pos: start})
		case strings.ContainsRune("(),;*=", runes[i]):
			tokens = append(tokens, token{kind: tokenSymbol, value: string(runes[i]), pos: start})
			i++
		default:
			return nil, newError(CodeSyntax, "unexpected character %q at position %d", runes[i], start+1)
		}
	}
	tokens = append(tokens, token{kind: tokenEOF, pos: len(runes)})
	return tokens, nil
}

// consumeExponent extends a numeric literal over a trailing exponent such as
// the "e-3" of 1.5e-3. It advances only when the full shape is present, so a
// stray "e" is left to lex as an identifier rather than truncating the number.
func consumeExponent(runes []rune, i int) int {
	if i >= len(runes) || (runes[i] != 'e' && runes[i] != 'E') {
		return i
	}
	next := i + 1
	if next < len(runes) && (runes[next] == '+' || runes[next] == '-') {
		next++
	}
	if next >= len(runes) || !unicode.IsDigit(runes[next]) {
		return i
	}
	for next < len(runes) && unicode.IsDigit(runes[next]) {
		next++
	}
	return next
}

type parser struct {
	tokens []token
	index  int
}

func (p *parser) parseStatement() (statement, error) {
	switch {
	case p.acceptKeyword("CREATE"):
		return p.parseCreateTable()
	case p.acceptKeyword("DROP"):
		return p.parseDropTable()
	case p.acceptKeyword("INSERT"):
		return p.parseInsert()
	case p.acceptKeyword("SELECT"):
		return p.parseSelect()
	case p.acceptKeyword("UPDATE"):
		return p.parseUpdate()
	case p.acceptKeyword("DELETE"):
		return p.parseDelete()
	default:
		return nil, p.syntaxError(p.peek(), "expected CREATE, DROP, INSERT, SELECT, UPDATE, or DELETE")
	}
}

func (p *parser) parseCreateTable() (statement, error) {
	if err := p.expectKeyword("TABLE"); err != nil {
		return nil, err
	}
	table, err := p.expectIdentifier("table name")
	if err != nil {
		return nil, err
	}
	if err := p.expectSymbol("("); err != nil {
		return nil, err
	}

	columns := make([]Column, 0, 4)
	for {
		name, err := p.expectIdentifier("column name")
		if err != nil {
			return nil, err
		}
		columnType, maxLength, err := p.parseDataType()
		if err != nil {
			return nil, err
		}
		column := Column{Name: name, Type: columnType, MaxLength: maxLength}
		for {
			switch {
			case p.acceptKeyword("PRIMARY"):
				if err := p.expectKeyword("KEY"); err != nil {
					return nil, err
				}
				column.PrimaryKey, column.Unique, column.NotNull = true, true, true
			case p.acceptKeyword("NOT"):
				if err := p.expectKeyword("NULL"); err != nil {
					return nil, err
				}
				column.NotNull = true
			case p.acceptKeyword("UNIQUE"):
				column.Unique = true
			default:
				columns = append(columns, column)
				goto constraintsDone
			}
		}
	constraintsDone:
		if p.acceptSymbol(")") {
			break
		}
		if err := p.expectSymbol(","); err != nil {
			return nil, err
		}
	}
	if len(columns) == 0 {
		return nil, p.syntaxError(p.peek(), "a table requires at least one column")
	}
	return createTableStatement{table: table, columns: columns}, nil
}

func (p *parser) parseDataType() (DataType, int, error) {
	tok := p.peek()
	if tok.kind != tokenIdentifier {
		return "", 0, p.syntaxError(tok, "expected column type")
	}
	p.index++
	switch strings.ToUpper(tok.value) {
	case "INTEGER", "INT":
		return TypeInteger, 0, nil
	case "TEXT":
		return TypeText, 0, nil
	case "VARCHAR":
		if err := p.expectSymbol("("); err != nil {
			return "", 0, err
		}
		lengthToken := p.peek()
		if lengthToken.kind != tokenNumber || strings.Contains(lengthToken.value, ".") || strings.HasPrefix(lengthToken.value, "-") {
			return "", 0, p.syntaxError(lengthToken, "expected a positive VARCHAR length")
		}
		var length int
		if _, err := fmt.Sscanf(lengthToken.value, "%d", &length); err != nil || length <= 0 {
			return "", 0, p.syntaxError(lengthToken, "expected a positive VARCHAR length")
		}
		p.index++
		if err := p.expectSymbol(")"); err != nil {
			return "", 0, err
		}
		return TypeText, length, nil
	case "BOOLEAN", "BOOL":
		return TypeBoolean, 0, nil
	case "FLOAT", "DOUBLE":
		return TypeFloat, 0, nil
	default:
		return "", 0, p.syntaxError(tok, "unsupported column type %q", tok.value)
	}
}

func (p *parser) parseDropTable() (statement, error) {
	if err := p.expectKeyword("TABLE"); err != nil {
		return nil, err
	}
	table, err := p.expectIdentifier("table name")
	return dropTableStatement{table: table}, err
}

func (p *parser) parseInsert() (statement, error) {
	if err := p.expectKeyword("INTO"); err != nil {
		return nil, err
	}
	table, err := p.expectIdentifier("table name")
	if err != nil {
		return nil, err
	}
	var columns []string
	if p.acceptSymbol("(") {
		columns, err = p.parseIdentifierList()
		if err != nil {
			return nil, err
		}
		if err := p.expectSymbol(")"); err != nil {
			return nil, err
		}
	}
	if err := p.expectKeyword("VALUES"); err != nil {
		return nil, err
	}
	if err := p.expectSymbol("("); err != nil {
		return nil, err
	}
	values, err := p.parseLiteralList()
	if err != nil {
		return nil, err
	}
	if err := p.expectSymbol(")"); err != nil {
		return nil, err
	}
	return insertStatement{table: table, columns: columns, values: values}, nil
}

func (p *parser) parseSelect() (statement, error) {
	stmt := selectStatement{}
	if p.acceptSymbol("*") {
		stmt.allColumns = true
	} else {
		columns, err := p.parseIdentifierList()
		if err != nil {
			return nil, err
		}
		stmt.columns = columns
	}
	if err := p.expectKeyword("FROM"); err != nil {
		return nil, err
	}
	table, err := p.expectIdentifier("table name")
	if err != nil {
		return nil, err
	}
	stmt.table = table
	stmt.where, err = p.parseOptionalWhere()
	return stmt, err
}

func (p *parser) parseUpdate() (statement, error) {
	table, err := p.expectIdentifier("table name")
	if err != nil {
		return nil, err
	}
	if err := p.expectKeyword("SET"); err != nil {
		return nil, err
	}
	assignments := make([]assignment, 0, 2)
	for {
		column, err := p.expectIdentifier("column name")
		if err != nil {
			return nil, err
		}
		if err := p.expectSymbol("="); err != nil {
			return nil, err
		}
		value, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment{column: column, value: value})
		if !p.acceptSymbol(",") {
			break
		}
	}
	where, err := p.parseOptionalWhere()
	return updateStatement{table: table, assignments: assignments, where: where}, err
}

func (p *parser) parseDelete() (statement, error) {
	if err := p.expectKeyword("FROM"); err != nil {
		return nil, err
	}
	table, err := p.expectIdentifier("table name")
	if err != nil {
		return nil, err
	}
	where, err := p.parseOptionalWhere()
	return deleteStatement{table: table, where: where}, err
}

func (p *parser) parseOptionalWhere() (*predicate, error) {
	if !p.acceptKeyword("WHERE") {
		return nil, nil
	}
	column, err := p.expectIdentifier("column name")
	if err != nil {
		return nil, err
	}
	if p.acceptKeyword("IS") {
		op := predicateIsNull
		if p.acceptKeyword("NOT") {
			op = predicateIsNotNull
		}
		if err := p.expectKeyword("NULL"); err != nil {
			return nil, err
		}
		return &predicate{column: column, op: op}, nil
	}
	if err := p.expectSymbol("="); err != nil {
		return nil, err
	}
	value, err := p.parseLiteral()
	if err != nil {
		return nil, err
	}
	return &predicate{column: column, op: predicateEqual, value: value}, nil
}

func (p *parser) parseIdentifierList() ([]string, error) {
	identifiers := make([]string, 0, 2)
	for {
		identifier, err := p.expectIdentifier("identifier")
		if err != nil {
			return nil, err
		}
		identifiers = append(identifiers, identifier)
		if !p.acceptSymbol(",") {
			return identifiers, nil
		}
	}
}

func (p *parser) parseLiteralList() ([]sqlLiteral, error) {
	literals := make([]sqlLiteral, 0, 2)
	for {
		literal, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		literals = append(literals, literal)
		if !p.acceptSymbol(",") {
			return literals, nil
		}
	}
}

func (p *parser) parseLiteral() (sqlLiteral, error) {
	tok := p.peek()
	p.index++
	switch tok.kind {
	case tokenString:
		return sqlLiteral{kind: literalString, value: tok.value}, nil
	case tokenNumber:
		return sqlLiteral{kind: literalNumber, value: tok.value}, nil
	case tokenIdentifier:
		switch strings.ToUpper(tok.value) {
		case "TRUE", "FALSE":
			return sqlLiteral{kind: literalBoolean, value: tok.value}, nil
		case "NULL":
			return sqlLiteral{kind: literalNull}, nil
		}
	}
	p.index--
	return sqlLiteral{}, p.syntaxError(tok, "expected a string, number, boolean, or NULL")
}

func (p *parser) peek() token {
	return p.tokens[p.index]
}

func (p *parser) acceptKeyword(keyword string) bool {
	tok := p.peek()
	if tok.kind == tokenIdentifier && strings.EqualFold(tok.value, keyword) {
		p.index++
		return true
	}
	return false
}

func (p *parser) expectKeyword(keyword string) error {
	if p.acceptKeyword(keyword) {
		return nil
	}
	return p.syntaxError(p.peek(), "expected %s", keyword)
}

func (p *parser) acceptSymbol(symbol string) bool {
	tok := p.peek()
	if tok.kind == tokenSymbol && tok.value == symbol {
		p.index++
		return true
	}
	return false
}

func (p *parser) expectSymbol(symbol string) error {
	if p.acceptSymbol(symbol) {
		return nil
	}
	return p.syntaxError(p.peek(), "expected %q", symbol)
}

func (p *parser) expectIdentifier(label string) (string, error) {
	tok := p.peek()
	if tok.kind != tokenIdentifier {
		return "", p.syntaxError(tok, "expected %s", label)
	}
	p.index++
	return strings.ToLower(tok.value), nil
}

func (p *parser) syntaxError(tok token, format string, args ...any) error {
	return newError(CodeSyntax, "%s at position %d", fmt.Sprintf(format, args...), tok.pos+1)
}
