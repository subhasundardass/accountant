package expr

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Context supplies the values an expression can read.
// A runtime or application can implement this interface.
type Context interface {
	Field(name string) (Value, bool)
	Variable(name string) (Value, bool)
	Formula(name string) (string, bool)
	CallFunction(name string, args []Value) (Value, error)
}

// Evaluate parses and evaluates a GoDL expression.
func Evaluate(source string, ctx Context) (Value, error) {
	tokens, err := tokenize(source)
	if err != nil {
		return nil, err
	}

	p := &expressionParser{
		tokens: tokens,
		ctx:    ctx,
	}

	value, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	if p.peek().kind != tokenEOF {
		return nil, fmt.Errorf("unexpected token %q", p.peek().text)
	}

	return value, nil
}

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenNumber
	tokenString
	tokenIdentifier
	tokenField
	tokenVariable
	tokenFormula
	tokenFunction
	tokenOperator
	tokenLeftParen
	tokenRightParen
	tokenComma
)

type token struct {
	kind tokenKind
	text string
}

// tokenize converts an expression into tokens.
func tokenize(source string) ([]token, error) {
	var tokens []token

	for i := 0; i < len(source); {
		r, size := utf8.DecodeRuneInString(source[i:])
		if r == utf8.RuneError && size == 1 {
			return nil, fmt.Errorf("invalid UTF-8 at position %d", i)
		}

		if unicode.IsSpace(r) {
			i += size
			continue
		}

		// String literals. A doubled quote represents one quote.
		if source[i] == '"' {
			i++

			var b strings.Builder
			closed := false

			for i < len(source) {
				if source[i] == '"' {
					if i+1 < len(source) && source[i+1] == '"' {
						b.WriteByte('"')
						i += 2
						continue
					}

					i++
					closed = true
					break
				}

				b.WriteByte(source[i])
				i++
			}

			if !closed {
				return nil, fmt.Errorf("unterminated string literal")
			}

			tokens = append(tokens, token{
				kind: tokenString,
				text: b.String(),
			})
			continue
		}

		// References:
		// $Amount       -> field
		// ##Discount    -> variable
		// $$Double(25)  -> function
		if strings.HasPrefix(source[i:], "##") {
			i += 2
			name, next, err := readName(source, i)
			if err != nil {
				return nil, fmt.Errorf("invalid variable reference: %w", err)
			}

			tokens = append(tokens, token{
				kind: tokenVariable,
				text: name,
			})
			i = next
			continue
		}

		if strings.HasPrefix(source[i:], "$$") {
			i += 2
			name, next, err := readName(source, i)
			if err != nil {
				return nil, fmt.Errorf("invalid function reference: %w", err)
			}

			tokens = append(tokens, token{
				kind: tokenFunction,
				text: name,
			})
			i = next
			continue
		}

		if source[i] == '$' {
			i++
			name, next, err := readName(source, i)
			if err != nil {
				return nil, fmt.Errorf("invalid field reference: %w", err)
			}

			tokens = append(tokens, token{
				kind: tokenField,
				text: name,
			})
			i = next
			continue
		}

		// Formula reference: @FullName
		if source[i] == '@' {
			i++
			name, next, err := readName(source, i)
			if err != nil {
				return nil, fmt.Errorf("invalid formula reference: %w", err)
			}

			tokens = append(tokens, token{
				kind: tokenFormula,
				text: name,
			})
			i = next
			continue
		}

		// Numeric literals, including decimal values.
		if isDigit(source[i]) ||
			(source[i] == '.' &&
				i+1 < len(source) &&
				isDigit(source[i+1])) {

			start := i
			dots := 0

			for i < len(source) &&
				(isDigit(source[i]) || source[i] == '.') {
				if source[i] == '.' {
					dots++
				}
				i++
			}

			if dots > 1 {
				return nil, fmt.Errorf("invalid number %q", source[start:i])
			}

			// Optional scientific notation, e.g. 1.5e3.
			if i < len(source) && (source[i] == 'e' || source[i] == 'E') {
				i++

				if i < len(source) && (source[i] == '+' || source[i] == '-') {
					i++
				}

				exponentStart := i
				for i < len(source) && isDigit(source[i]) {
					i++
				}

				if exponentStart == i {
					return nil, fmt.Errorf("invalid number %q", source[start:i])
				}
			}

			number := source[start:i]
			if _, err := strconv.ParseFloat(number, 64); err != nil {
				return nil, fmt.Errorf("invalid number %q: %w", number, err)
			}

			tokens = append(tokens, token{
				kind: tokenNumber,
				text: number,
			})
			continue
		}

		// Identifiers and logical keywords.
		if isIdentifierStart(source[i]) {
			start := i
			for i < len(source) && isIdentifierCharacter(source[i]) {
				i++
			}

			tokens = append(tokens, token{
				kind: tokenIdentifier,
				text: source[start:i],
			})
			continue
		}

		switch source[i] {
		case '(':
			tokens = append(tokens, token{
				kind: tokenLeftParen,
				text: "(",
			})
			i++
			continue

		case ')':
			tokens = append(tokens, token{
				kind: tokenRightParen,
				text: ")",
			})
			i++
			continue

		case ',':
			tokens = append(tokens, token{
				kind: tokenComma,
				text: ",",
			})
			i++
			continue
		}

		// Two-character comparison operators.
		if i+1 < len(source) {
			op := source[i : i+2]

			switch op {
			case "<=", ">=", "<>", "!=", "==":
				tokens = append(tokens, token{
					kind: tokenOperator,
					text: op,
				})
				i += 2
				continue
			}
		}

		// Single-character operators.
		if strings.ContainsRune("+-*/=<>", rune(source[i])) {
			tokens = append(tokens, token{
				kind: tokenOperator,
				text: source[i : i+1],
			})
			i++
			continue
		}

		return nil, fmt.Errorf("unexpected character %q", source[i])
	}

	return append(tokens, token{kind: tokenEOF}), nil
}

// readName reads a field, variable, formula, or function name.
// Hyphens are deliberately excluded so that subtraction works:
//
//	$Amount - $Discount
func readName(source string, start int) (string, int, error) {
	i := start

	if i >= len(source) || !isIdentifierStart(source[i]) {
		return "", start, fmt.Errorf("expected a name")
	}

	for i < len(source) && isNameCharacter(source[i]) {
		i++
	}

	return source[start:i], i, nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isIdentifierStart(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z')
}

func isIdentifierCharacter(b byte) bool {
	return isIdentifierStart(b) || isDigit(b)
}

func isNameCharacter(b byte) bool {
	return isIdentifierCharacter(b) || b == '.'
}

type expressionParser struct {
	tokens []token
	pos    int
	ctx    Context
	depth  int
}

const maxExpressionDepth = 32

func (p *expressionParser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{kind: tokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *expressionParser) take() token {
	t := p.peek()

	if p.pos < len(p.tokens) {
		p.pos++
	}

	return t
}

func (p *expressionParser) accept(kind tokenKind, value string) bool {
	t := p.peek()

	if t.kind != kind {
		return false
	}

	if value != "" && !strings.EqualFold(t.text, value) {
		return false
	}

	p.take()
	return true
}

// parseExpression uses precedence climbing.
//
// Precedence, from low to high:
//
//	or
//	and
//	comparison
//	addition / subtraction
//	multiplication / division
func (p *expressionParser) parseExpression(minPrecedence int) (Value, error) {
	if p.depth >= maxExpressionDepth {
		return nil, fmt.Errorf("expression nesting limit exceeded")
	}

	p.depth++
	defer func() {
		p.depth--
	}()

	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		op := p.peek()
		precedence := operatorPrecedence(op)

		if precedence < minPrecedence {
			break
		}

		p.take()

		// Short-circuit AND.
		if strings.EqualFold(op.text, "and") && !ToBool(left) {
			if _, err := p.parseExpression(precedence + 1); err != nil {
				return nil, err
			}

			left = false
			continue
		}

		// Short-circuit OR.
		if strings.EqualFold(op.text, "or") && ToBool(left) {
			if _, err := p.parseExpression(precedence + 1); err != nil {
				return nil, err
			}

			left = true
			continue
		}

		right, err := p.parseExpression(precedence + 1)
		if err != nil {
			return nil, err
		}

		left, err = applyOperator(op.text, left, right)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func operatorPrecedence(t token) int {
	if t.kind == tokenIdentifier {
		switch strings.ToLower(t.text) {
		case "or":
			return 1
		case "and":
			return 2
		}
		return -1
	}

	if t.kind != tokenOperator {
		return -1
	}

	switch t.text {
	case "=", "==", "!=", "<>", "<", "<=", ">", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/":
		return 5
	}

	return -1
}

func (p *expressionParser) parseUnary() (Value, error) {
	t := p.peek()

	if t.kind == tokenIdentifier &&
		strings.EqualFold(t.text, "not") {
		p.take()

		value, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		return !ToBool(value), nil
	}

	if t.kind == tokenOperator && (t.text == "-" || t.text == "+") {
		p.take()

		value, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		n, ok := ToNumber(value)
		if !ok {
			return nil, fmt.Errorf(
				"cannot apply %s to %q",
				t.text,
				ToString(value),
			)
		}

		if t.text == "-" {
			n = -n
		}

		return n, nil
	}

	return p.parsePrimary()
}

func (p *expressionParser) parsePrimary() (Value, error) {
	t := p.take()

	switch t.kind {
	case tokenNumber:
		return strconv.ParseFloat(t.text, 64)

	case tokenString:
		return t.text, nil

	case tokenLeftParen:
		value, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if !p.accept(tokenRightParen, ")") {
			return nil, fmt.Errorf("expected closing parenthesis")
		}

		return value, nil

	case tokenField:
		if p.ctx == nil {
			return nil, fmt.Errorf(
				"cannot resolve field $%s without an expression context",
				t.text,
			)
		}

		value, ok := p.ctx.Field(t.text)
		if !ok {
			return nil, fmt.Errorf("unknown field $%s", t.text)
		}

		return Normalize(value), nil

	case tokenVariable:
		if p.ctx == nil {
			return nil, fmt.Errorf(
				"cannot resolve variable ##%s without an expression context",
				t.text,
			)
		}

		value, ok := p.ctx.Variable(t.text)
		if !ok {
			return nil, fmt.Errorf("unknown variable ##%s", t.text)
		}

		return Normalize(value), nil

	case tokenFormula:
		if p.ctx == nil {
			return nil, fmt.Errorf(
				"cannot resolve formula @%s without an expression context",
				t.text,
			)
		}

		source, ok := p.ctx.Formula(t.text)
		if !ok {
			return nil, fmt.Errorf("unknown formula @%s", t.text)
		}

		return Evaluate(source, p.ctx)

	case tokenFunction:
		return p.parseFunction(t.text)

	case tokenIdentifier:
		switch strings.ToLower(t.text) {
		case "yes", "true":
			return true, nil

		case "no", "false":
			return false, nil

		case "nil", "null":
			return nil, nil

		case "if":
			return p.parseIf()
		}

		return nil, fmt.Errorf("unknown identifier %q", t.text)

	default:
		return nil, fmt.Errorf("unexpected token %q", t.text)
	}
}

// parseFunction supports:
//
//	$$Double(25)
//	$$Add(10, 20)
//	$$Format($Amount, "0.00")
func (p *expressionParser) parseFunction(name string) (Value, error) {
	if p.ctx == nil {
		return nil, fmt.Errorf(
			"cannot call $$%s without an expression context",
			name,
		)
	}

	if !p.accept(tokenLeftParen, "(") {
		return nil, fmt.Errorf("expected '(' after function $$%s", name)
	}

	var args []Value

	// Empty argument list.
	if p.accept(tokenRightParen, ")") {
		return p.ctx.CallFunction(name, args)
	}

	for {
		arg, err := p.parseExpression(0)
		if err != nil {
			return nil, fmt.Errorf("function $$%s: %w", name, err)
		}

		args = append(args, arg)

		if p.accept(tokenRightParen, ")") {
			break
		}

		if !p.accept(tokenComma, ",") {
			return nil, fmt.Errorf(
				"expected ',' or ')' in arguments of $$%s",
				name,
			)
		}
	}

	return p.ctx.CallFunction(name, args)
}

// parseIf supports:
//
//	if $Amount > 100 then "High" else "Low"
//
// The optional endif keyword is accepted:
//
//	if $Amount > 100 then "High" else "Low" endif
func (p *expressionParser) parseIf() (Value, error) {
	condition, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	if !p.accept(tokenIdentifier, "then") {
		return nil, fmt.Errorf("expected then in conditional expression")
	}

	// Parse the selected and unselected branches.
	// Both branches are parsed to keep the token stream valid.
	// Only the selected branch's value is returned.
	if ToBool(condition) {
		selected, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if !p.accept(tokenIdentifier, "else") {
			return nil, fmt.Errorf("expected else in conditional expression")
		}

		if _, err := p.parseExpression(0); err != nil {
			return nil, err
		}

		p.accept(tokenIdentifier, "endif")
		return selected, nil
	}

	if _, err := p.parseExpression(0); err != nil {
		return nil, err
	}

	if !p.accept(tokenIdentifier, "else") {
		return nil, fmt.Errorf("expected else in conditional expression")
	}

	selected, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	p.accept(tokenIdentifier, "endif")
	return selected, nil
}

func applyOperator(op string, left, right Value) (Value, error) {
	switch strings.ToLower(op) {
	case "and":
		return ToBool(left) && ToBool(right), nil

	case "or":
		return ToBool(left) || ToBool(right), nil

	case "=", "==":
		return compare(left, right) == 0, nil

	case "!=", "<>":
		return compare(left, right) != 0, nil

	case "<":
		return compare(left, right) < 0, nil

	case "<=":
		return compare(left, right) <= 0, nil

	case ">":
		return compare(left, right) > 0, nil

	case ">=":
		return compare(left, right) >= 0, nil

	case "+":
		// String concatenation.
		if _, ok := left.(string); ok {
			return ToString(left) + ToString(right), nil
		}

		if _, ok := right.(string); ok {
			return ToString(left) + ToString(right), nil
		}

		a, oka := ToNumber(left)
		b, okb := ToNumber(right)

		if !oka || !okb {
			return nil, fmt.Errorf("cannot add %T and %T", left, right)
		}

		return a + b, nil

	case "-":
		a, oka := ToNumber(left)
		b, okb := ToNumber(right)

		if !oka || !okb {
			return nil, fmt.Errorf("operator - requires numeric values")
		}

		return a - b, nil

	case "*":
		a, oka := ToNumber(left)
		b, okb := ToNumber(right)

		if !oka || !okb {
			return nil, fmt.Errorf("operator * requires numeric values")
		}

		return a * b, nil

	case "/":
		a, oka := ToNumber(left)
		b, okb := ToNumber(right)

		if !oka || !okb {
			return nil, fmt.Errorf("operator / requires numeric values")
		}

		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}

		return a / b, nil
	}

	return nil, fmt.Errorf("unsupported operator %q", op)
}
