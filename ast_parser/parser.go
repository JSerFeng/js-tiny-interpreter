package ast_parser

import (
	"jsInterpreter/lexer"
	"jsInterpreter/logger"
	"strings"
	"unicode/utf8"
)

type AstParser struct {
	RawSource string
	Tokens    []lexer.Token
	Curr      int
	HasError  bool
	Log       logger.Logger
}

type AstError struct {
	Loc logger.Loc
	Msg string
}

func NewParser(source string) AstParser {
	l := lexer.NewLexer(source)
	p := AstParser{
		RawSource: source,
		Tokens:    l.Tokenizer(),
		Curr:      0,
		HasError:  false,
		Log:       logger.Logger{Content: source},
	}

	if l.HasError {
		p.HasError = true
	}
	return p
}

func (p *AstParser) Error(err AstError) {
	p.HasError = true
	panic(err)
}

func (p *AstParser) CurToken() lexer.Token {
	if p.IsEnd() {
		return lexer.Token{lexer.TEndOfFile, p.Prev().Loc}
	}
	return p.Tokens[p.Curr]
}

func (p *AstParser) IsEnd() bool {
	return p.Curr >= len(p.Tokens)
}

func (p *AstParser) Step() lexer.Token {
	if p.Curr == len(p.Tokens)-1 {
		defer func() {
			p.Curr++
		}()
		return p.Tokens[p.Curr]
	}
	t := p.Tokens[p.Curr]
	p.Curr++
	return t
}

func (p *AstParser) Consume(t lexer.T) lexer.Token {
	token := p.CurToken()
	if token.T == t {
		return p.Step()
	} else {
		panic(AstError{
			token.Loc,
			"expect " + lexer.Token2StringMap[t] + ", but found " + lexer.Raw(token.Loc, p.RawSource)})
	}
}

func (p *AstParser) Raw(t lexer.Token) string {
	return lexer.Raw(t.Loc, p.RawSource)
}

func (p *AstParser) Check(matches ...lexer.T) bool {
	if p.IsEnd() {
		return false
	}
	currToken := p.CurToken()
	for _, t := range matches {
		if currToken.T == t {
			return true
		}
	}
	return false
}

func (p *AstParser) Prev() *lexer.Token {
	return &p.Tokens[p.Curr-1]
}

func (p *AstParser) AdvanceToNextStmt() {

}

func (p *AstParser) Parse() Stmt {
	defer func() {
		err := recover()
		if err != nil {
			p.HasError = true
			if err, isErr := err.(AstError); isErr {
				p.Log.Print(logger.LError, err.Loc, err.Msg)
			} else {
				panic(err)
			}
		}
	}()
	loc := logger.Loc{Line: 1}
	stmts := make([]*Stmt, 0)
	for !p.IsEnd() {
		stmt := p.stmt()
		stmts = append(stmts, &stmt)
	}
	p.calcLocFromPrevToken(&loc)
	return Stmt{
		Loc:  loc,
		Data: &SProgram{Body: stmts},
	}
}

func (p *AstParser) calcLocFromPrevToken(loc *logger.Loc) {
	prevLoc := p.Prev().Loc
	loc.Len = prevLoc.Offset + prevLoc.Len - loc.Offset
}

/*
解析语句
stmt ->
	SFunctionDecl
	SBlock
	SExpr
*/
func (p *AstParser) stmt() Stmt {
	token := p.CurToken()
	var stmt Stmt
	switch token.T {
	//将function看成表达式而不是语句
	//case lexer.TFunction:
	//	loc := token.Loc
	//	/*消耗掉function关键字*/
	//	p.Step()
	//	funcStmt := p.sFuncDecl()
	//	p.calcLocFromPrevToken(&loc)
	//	funcStmt.Loc = loc
	//	stmt = funcStmt
	case lexer.TOpenBrace:
		stmt = p.sBlock()
	case lexer.TClass:
		stmt = p.class()
	case lexer.TLet, lexer.TConst, lexer.TVar:
		stmt = p.sVarDecl()
	case lexer.TSemicolon:
		//防止;开头
		p.Step()
		stmt = Stmt{token.Loc, nil}
	case lexer.TFor:
		stmt = p.sFor()
	case lexer.TIf:
		stmt = p.sCondition()
	case lexer.TBreak, lexer.TContinue:
		p.Step()
		if token.T == lexer.TBreak {
			stmt = Stmt{token.Loc, &SBreak{}}
		} else {
			stmt = Stmt{token.Loc, &SContinue{}}
		}
	case lexer.TReturn:
		stmt = p.sReturn()
	default:
		expr := p.expr()
		stmt = Stmt{
			Loc:  expr.Loc,
			Data: &SExpr{Expr: expr},
		}
	}

	//每条语句后如果有分号，跳过它
	if p.Check(lexer.TSemicolon) {
		p.Step()
	}

	return stmt
}

func (p *AstParser) sVarDecl() Stmt {
	token := p.Step()
	loc := token.Loc
	name := p.Consume(lexer.TIdentifier)
	var kind VarKind
	switch token.T {
	case lexer.TLet:
		kind = VLet
	case lexer.TConst:
		kind = VConst
	default:
		kind = VVar
	}

	varDecl := p.varDecl(EIdentifier{Value: p.Raw(name)}, kind)
	p.calcLocFromPrevToken(&loc)
	return Stmt{
		Loc:  loc,
		Data: &varDecl,
	}
}

/*
为了方便class中声明变量复用这个方法，假设let，const以及变量名已经被consume了
@Param id 变量名
*/
func (p *AstParser) varDecl(id EIdentifier, kind VarKind) SVarDecl {
	var init *Expr = nil
	if p.Check(lexer.TEquals) {
		p.Step()
		initExpr := p.expr()
		init = &initExpr
	}

	return SVarDecl{
		Id:   id,
		Init: init,
	}
}

func (p *AstParser) sReturn() Stmt {
	loc := p.CurToken().Loc
	p.Consume(lexer.TReturn)

	expr := p.expr()
	p.calcLocFromPrevToken(&loc)
	return Stmt{loc, &SReturn{expr}}
}

/*
函数既可以是一个语句，也可以是表达式 eg: arr.map(function mapFn() {})
这里假设function的token已经被消耗掉了, 因为可以让后面class中的方法声明复用
function a() { }
*/
func (p *AstParser) sFuncDecl() Stmt {
	loc := p.CurToken().Loc
	var id *EIdentifier = nil
	if p.Check(lexer.TIdentifier) {
		id = &EIdentifier{Value: p.Raw(p.Step())}
	} else {
		id = &EIdentifier{"Annoymous"}
	}
	fnDecl := p.funcDecl(id)

	p.calcLocFromPrevToken(&loc)

	return Stmt{
		Loc:  loc,
		Data: &fnDecl,
	}
}

/*
@Params id 函数名
解析function add() { }
               ^^^^^^
*/
func (p *AstParser) funcDecl(id *EIdentifier) SFunctionDecl {
	loc := p.Consume(lexer.TOpenParen).Loc
	//函数形参
	params := make([]EIdentifier, 0)
Params:
	for !p.IsEnd() {
		param := p.CurToken()
		switch param.T {
		case lexer.TCloseParen:
			break Params
		case lexer.TComma:
			p.Step()
			p.Error(AstError{param.Loc, "Syntax Error: can't allow comma ahead of parameter"})
		case lexer.TIdentifier:
			paramId := EIdentifier{Value: lexer.Raw(param.Loc, p.RawSource)}
			params = append(params, paramId)
			p.Step()
			if p.CurToken().T == lexer.TComma {
				p.Step()
			}
		}
	}
	if p.IsEnd() {
		p.Error(AstError{p.CurToken().Loc, "missing close parentheses"})
	}
	p.Consume(lexer.TCloseParen)

	body := p.body()
	p.calcLocFromPrevToken(&loc)
	return SFunctionDecl{
		Id:     id,
		Params: params,
		Body:   &body,
	}
}

/*
class Child extends Father {
	field = 1;
	sayHello() { }
}
*/
func (p *AstParser) class() Stmt {
	loc := p.CurToken().Loc
	p.Consume(lexer.TClass)
	id := p.Consume(lexer.TIdentifier)
	className := lexer.Raw(id.Loc, p.RawSource)
	var superClass *Expr = nil
	maybeExtends := p.CurToken()
	if maybeExtends.T == lexer.TExtends {
		p.Step()
		expr := p.expr()
		superClass = &expr
	}

	//类内部域
	classMembers := make([]ClassMember, 0)
	p.Consume(lexer.TOpenBrace)
ClassFields:
	for !p.IsEnd() {
		token := p.CurToken()
		switch token.T {
		case lexer.TCloseBrace:
			break ClassFields
		case lexer.TIdentifier:
			memberLoc := token.Loc
			id := EIdentifier{Value: p.Raw(p.Step())}
			parenOrEquals := p.CurToken()
			switch parenOrEquals.T {
			case lexer.TOpenParen:
				// 类的方法class X { method() { } }
				fnBody := p.funcDecl(&id)
				//结尾位置 计算整条语句位置
				endPos := p.Prev().Loc.Offset + p.Prev().Loc.Len
				memberLoc.Len = endPos - memberLoc.Offset
				classMembers = append(classMembers, ClassMember{
					Kind:  MethodDefinition,
					Loc:   memberLoc,
					Key:   id,
					Value: &fnBody,
				})
			case lexer.TEquals:
				// 类的属性class X { name = 1 }
				propBody := p.varDecl(id, VLet)
				//结尾位置 计算整条语句位置
				endPos := p.Prev().Loc.Offset + p.Prev().Loc.Len
				memberLoc.Len = endPos - memberLoc.Offset
				classMembers = append(classMembers, ClassMember{
					Kind:  PropertyDefinition,
					Loc:   memberLoc,
					Key:   id,
					Value: &propBody,
				})
			}
		default:
			p.Error(AstError{token.Loc, "Syntax Error: unexpected error"})
		}
	}

	if p.IsEnd() {
		p.Error(AstError{loc, "missing closing brace"})
	}

	p.Consume(lexer.TCloseBrace)

	return Stmt{
		Loc: loc,
		Data: &SClass{
			SuperClass: superClass,
			Id:         EIdentifier{Value: className},
			Body:       nil,
		},
	}
}

func (p *AstParser) sBlock() Stmt {
	loc := p.CurToken().Loc
	block := p.block()
	p.calcLocFromPrevToken(&loc)
	return Stmt{
		Loc:  loc,
		Data: &block,
	}
}

func (p *AstParser) body() SBody {
	loc := p.CurToken().Loc
	block := p.block()
	p.calcLocFromPrevToken(&loc)
	return SBody{
		Loc:  loc,
		Data: &block,
	}
}

func (p *AstParser) block() SBlock {
	p.Consume(lexer.TOpenBrace)

	stmts := make([]*Stmt, 0)

	for p.CurToken().T != lexer.TCloseBrace && !p.IsEnd() {
		stmt := p.stmt()
		stmts = append(stmts, &stmt)
	}

	if p.IsEnd() {
		p.Error(AstError{
			Loc: p.CurToken().Loc,
			Msg: "missing close brace",
		})
	}
	p.Consume(lexer.TCloseBrace)
	return SBlock{
		Data: stmts,
	}
}

/*
for (initialize; situation; changes) {}
*/
func (p *AstParser) sFor() Stmt {
	forKeyWord := p.Consume(lexer.TFor)
	p.Consume(lexer.TOpenParen)
	loc := forKeyWord.Loc
	initializer := p.stmt()

	condition := p.expr()
	p.Consume(lexer.TSemicolon)

	reset := p.stmt()

	p.Consume(lexer.TCloseParen)
	body := p.body()
	p.calcLocFromPrevToken(&loc)
	return Stmt{
		Loc: loc,
		Data: &SFor{
			Initializer: &initializer,
			Condition:   &condition,
			Reset:       &reset,
			Body:        &body,
		},
	}
}

/*
if (condition) {} else if (condition) {}
*/
func (p *AstParser) sCondition() Stmt {
	ifToken := p.Consume(lexer.TIf)
	loc := ifToken.Loc

	p.Consume(lexer.TOpenParen)
	condition := p.expr()
	p.Consume(lexer.TCloseParen)
	body := p.body()

	branches := []Branch{Branch{&condition, &body}}
Branch:
	for p.Check(lexer.TElse) {
		p.Step()
		switch p.CurToken().T {
		case lexer.TIf:
			p.Step()
			//else-if
			p.Consume(lexer.TOpenParen)
			_condition := p.expr()
			p.Consume(lexer.TCloseParen)
			_body := p.body()
			branches = append(branches, Branch{
				Condition: &_condition,
				Body:      &_body,
			})
		default:
			_body := p.body()
			branches = append(branches, Branch{
				Condition: nil,
				Body:      &_body,
			})
			break Branch
		}
	}

	p.calcLocFromPrevToken(&loc)
	return Stmt{
		Loc:  loc,
		Data: &SCondition{Branches: branches},
	}
}

/*
解析表达式优先级由上至下 从小到大
return
assignment
equals
or(||)
and(&&)
compare
binary
plusAndMinus(+ or -)
multiplyAndDivision(* or /)
unary
call
memberExpr
paren
primary
*/
func (p *AstParser) expr() Expr {
	return p.assignment()
}

func (p *AstParser) assignment() Expr {
	if p.IsEnd() {
		return Expr{
			Loc:  logger.Loc{},
			Data: nil,
		}
	}
	loc := p.CurToken().Loc
	expr := p.equals()

	if p.Check(lexer.TEquals) {
		p.Step()
		assignment := p.assignment()
		p.calcLocFromPrevToken(&loc)
		target := expr
		expr = Expr{
			Loc: loc,
			Data: &EAssign{
				Target:     &target,
				Assignment: &assignment,
			},
		}
	}

	return expr
}

/**
equals -> or (== equals)*
*/
func (p *AstParser) equals() Expr {
	expr := p.or()
Equals:
	for {
		token := p.CurToken()
		switch token.T {
		case lexer.TEqualsEquals, lexer.TEqualsEqualsEquals:
			p.Step()
			equalTarget := p.equals()
			expr = Expr{
				Loc: logger.Loc{
					Offset: expr.Loc.Offset,
					Len:    expr.Loc.Len + token.Loc.Len + equalTarget.Loc.Len,
					Line:   expr.Loc.Line,
				},
				Data: &EBinary{
					Op:    EOp(token.T),
					Left:  expr,
					Right: equalTarget,
				},
			}
		default:
			break Equals
		}
	}

	return expr
}

/**
or -> and (&& or)*
*/
func (p *AstParser) or() Expr {
	expr := p.and()
Bar:
	for {
		token := p.CurToken()
		loc := token.Loc
		switch token.T {
		case lexer.TBarBar:
			p.Step()
			andTarget := p.or()
			p.calcLocFromPrevToken(&loc)

			expr = Expr{
				Loc: loc,
				Data: &EBinary{
					Op:    EOp(lexer.TBarBar),
					Left:  expr,
					Right: andTarget,
				},
			}
		default:
			break Bar
		}
	}
	return expr
}

/**
and -> compare (&& and)*
*/
func (p *AstParser) and() Expr {
	expr := p.compare()
Ampersand:
	for {
		token := p.CurToken()
		loc := token.Loc
		switch token.T {
		case lexer.TAmpersandAmpersand:
			p.Step()
			andTarget := p.and()
			p.calcLocFromPrevToken(&loc)
			expr = Expr{
				Loc: loc,
				Data: &EBinary{
					Op:    EOp(lexer.TAmpersandAmpersand),
					Left:  expr,
					Right: andTarget,
				},
			}
		default:
			break Ampersand
		}
	}
	return expr
}

/**
compare -> add (>= compare)*
*/
func (p *AstParser) compare() Expr {
	expr := p.plusAndMinus()
	Loc := expr.Loc

Compare:
	for {
		token := p.CurToken()
		switch token.T {
		case
			lexer.TLessThan, lexer.TLessThanEquals,
			lexer.TGreaterThan, lexer.TGreaterThanEquals:
			p.Step()
			target := p.compare()
			Loc.Len += token.Loc.Len + target.Loc.Len
			expr = Expr{
				Loc: Loc,
				Data: &EBinary{
					Op:    EOp(token.T),
					Left:  expr,
					Right: target,
				},
			}
		default:
			break Compare
		}
	}

	return expr
}

/*
add -> mul (+ add)*
*/
func (p *AstParser) plusAndMinus() Expr {
	token := p.CurToken()
	loc := token.Loc
	expr := p.multiplyAndDivision()

Plus:
	for {
		nextToken := p.CurToken()
		switch nextToken.T {
		case lexer.TPlus, lexer.TMinus:
			p.Step()
			right := p.plusAndMinus()
			p.calcLocFromPrevToken(&loc)
			expr = Expr{
				Loc: loc,
				Data: &EBinary{
					Op:    EOp(nextToken.T),
					Left:  expr,
					Right: right,
				},
			}
		default:
			break Plus
		}
	}
	return expr
}

/**
multiply -> unary (* multiply)*
*/
func (p *AstParser) multiplyAndDivision() Expr {
	token := p.CurToken()
	loc := token.Loc
	expr := p.unary()
Mul:
	for {
		nextToken := p.CurToken()
		switch nextToken.T {
		case lexer.TAsterisk, lexer.TSlash, lexer.TPercent:
			p.Step()
			right := p.multiplyAndDivision()
			p.calcLocFromPrevToken(&loc)
			expr = Expr{
				Loc: loc,
				Data: &EBinary{
					Op:    EOp(nextToken.T),
					Left:  expr,
					Right: right,
				},
			}
		default:
			break Mul
		}
	}
	return expr
}

/*
!a
-a
++a
a++
*/
func (p *AstParser) unary() Expr {
	token := p.CurToken()
	loc := token.Loc
	var expr Expr

	switch token.T {
	case lexer.TExclamation, lexer.TMinus, lexer.TPlusPlus, lexer.TMinusMinus:
		p.Step()
		expr = Expr{
			Loc: token.Loc,
			Data: &EUnary{
				Op:    EOp(token.T),
				Value: p.unary(),
			},
		}
	default:
		expr = p.callExpr()
	}

	//处理后置一元运算符，例如 num++
	if p.Check(lexer.TPlusPlus, lexer.TMinusMinus) {
		op := p.Step().T
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc: loc,
			Data: &EUnary{
				Op:          EOp(op),
				Value:       expr,
				Association: AssociationRight,
			},
		}
	}

	return expr
}

func (p *AstParser) callExpr() Expr {
	expr := p.memberExpr()

	token := p.CurToken()
	if token.T == lexer.TOpenParen {
		p.Step()
		loc := expr.Loc
		args := []*Expr{}
		for !p.IsEnd() && p.CurToken().T != lexer.TCloseParen {
			arg := p.expr()
			args = append(args, &arg)
			if p.Check(lexer.TComma) {
				p.Step()
			}
		}
		p.Consume(lexer.TCloseParen)
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc: loc,
			Data: &ECallExpr{
				Callee: expr,
				Args:   args,
			},
		}
	}

	return expr
}

/*
a.b.c对应
Obj: EMemberExpr{
	Obj: 		a,
	Property: b
}
Property: c
*/
func (p *AstParser) memberExpr() Expr {
	expr := p.paren()
	token := p.CurToken()
	if token.T == lexer.TDot {
		loc := expr.Loc
		p.Step()
		property := p.paren()
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc: loc,
			Data: &EMemberExpr{
				Obj:      expr,
				Property: EIdentifier{property.Data.(*EIdentifier).Value},
			},
		}

		for !p.IsEnd() && p.CurToken().T == lexer.TDot {
			p.Step()
			newProperty := p.paren()
			oldLoc := loc
			p.calcLocFromPrevToken(&loc)
			expr = Expr{
				Loc: loc,
				Data: &EMemberExpr{
					Obj: Expr{
						Loc:  oldLoc,
						Data: expr.Data,
					},
					Property: EIdentifier{newProperty.Data.(*EStringLiteral).Value},
				},
			}
		}
	}
	return expr
}

func (p *AstParser) paren() Expr {
	token := p.CurToken()
	loc := token.Loc
	var expr Expr

	switch token.T {
	case lexer.TOpenParen:
		p.Step()
		value := p.assignment()
		p.Consume(lexer.TCloseParen)
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc:  loc,
			Data: &EParen{Data: value},
		}
	default:
		expr = p.primary()
	}

	return expr
}

func (p *AstParser) primary() Expr {
	token := p.CurToken()
	loc := token.Loc
	var expr Expr
	switch token.T {
	case lexer.TNumericLiteral:
		name := lexer.Raw(token.Loc, p.RawSource)
		expr = Expr{
			Loc:  loc,
			Data: &ENumericLiteral{Value: name},
		}
	case lexer.TTrue:
		expr = Expr{
			Loc:  loc,
			Data: &EBoolLiteral{Value: true},
		}
	case lexer.TFalse:
		expr = Expr{
			Loc:  loc,
			Data: &EBoolLiteral{Value: false},
		}
	case lexer.TStringLiteral:
		name := lexer.Raw(token.Loc, p.RawSource)
		if (strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"")) ||
			(strings.HasPrefix(name, "'") && strings.HasSuffix(name, "'")) {
			prefix, _ := utf8.DecodeRuneInString(name)
			name = strings.TrimSuffix(strings.TrimPrefix(name, string(prefix)), string(prefix))
		} else {
			p.Error(AstError{token.Loc, "string literal should wrapped in ' or \" "})
		}
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc:  loc,
			Data: &EStringLiteral{Value: name},
		}
	case lexer.TIdentifier:
		name := lexer.Raw(token.Loc, p.RawSource)
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc:  loc,
			Data: &EIdentifier{Value: name},
		}
	case lexer.TFunction: //Function Expression
		p.Step()
		var id *EIdentifier = nil
		if p.CurToken().T == lexer.TIdentifier {
			id = &EIdentifier{lexer.Raw(p.CurToken().Loc, p.RawSource)}
			p.Consume(lexer.TIdentifier)
		}
		sFnDecl := p.funcDecl(id)
		p.calcLocFromPrevToken(&loc)
		return Expr{
			Loc: loc,
			Data: &EFunctionExpr{
				Id:     sFnDecl.Id,
				Params: sFnDecl.Params,
				Body:   sFnDecl.Body,
			},
		}
	case lexer.TEqualsGreaterThan: //arrow function
	case lexer.TOpenBrace: //object literal
		p.Step()
		loc := token.Loc
		properties := []*Expr{}
	ObjectLiteral:
		for !p.IsEnd() {
			var key Expr
			var val Expr
			token := p.CurToken()
			switch token.T {
			case lexer.TOpenBracket:
				p.Step()
				key = p.expr()
				p.Consume(lexer.TCloseBracket)
				p.Consume(lexer.TColon)
				val = p.expr()
				if p.CurToken().T == lexer.TComma {
					p.Consume(lexer.TComma)
				} else {
					break ObjectLiteral
				}
				properties = append(properties, &key, &val)
			case lexer.TCloseBrace:
				break ObjectLiteral
			case lexer.TIdentifier, lexer.TStringLiteral, lexer.TNumericLiteral:
				if token.T == lexer.TStringLiteral {
					key = p.primary()
				} else {
					key = Expr{token.Loc, &EStringLiteral{lexer.Raw(token.Loc, p.RawSource)}}
					p.Step()
				}
				p.Consume(lexer.TColon)
				val = p.expr()
				properties = append(properties, &key, &val)
				if p.CurToken().T == lexer.TComma {
					p.Consume(lexer.TComma)
				} else {
					break ObjectLiteral
				}
			}
		}
		p.Consume(lexer.TCloseBrace)
		p.calcLocFromPrevToken(&loc)
		return Expr{
			Loc: loc,
			Data: &EObjectLiteral{
				Proto:       nil,
				Constructor: nil,
				Properties:  properties,
			},
		}
	case lexer.TOpenBracket: // array literal
		p.Step()
		// [(expr(,)?)*]
		arr := []*Expr{}
		loc := token.Loc
		var length uint = 0
	Array:
		for !p.IsEnd() {
			curToken := p.CurToken()
			switch curToken.T {
			case lexer.TComma:
				p.Step()
			case lexer.TCloseBracket:
				break Array
			default:
				expr := p.expr()
				length++
				arr = append(arr, &expr)
			}
		}
		p.Consume(lexer.TCloseBracket)
		p.calcLocFromPrevToken(&loc)
		return Expr{
			Loc: loc,
			Data: &EArrayLiteral{
				Arr:    arr,
				Length: length,
			},
		}
	default:
		p.Step()
		p.Error(AstError{token.Loc, "unexpected token"})
	}
	p.Step()

	curToken := p.CurToken()

	//a[0]
	if curToken.T == lexer.TOpenBracket {
		p.Step()
		idx := p.expr()
		p.Consume(lexer.TCloseBracket)
		p.calcLocFromPrevToken(&loc)
		expr = Expr{
			Loc: loc,
			Data: &EIndex{
				Target: expr,
				Idx:    idx,
			},
		}
	}

	return expr
}
