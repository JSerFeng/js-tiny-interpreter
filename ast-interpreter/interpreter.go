package ast_interpreter

import (
	"fmt"
	"jsInterpreter/ast_parser"
	"jsInterpreter/logger"
	"strconv"
)

func (s *InterpreterStat) Call(f *JsFunction, args []JsValue) (returnValue *JsValue) {
	oldScope := s.Scope
	returnValue = &JsValue{nil, Undefined}
	defer func() {
		s.Scope = oldScope
		err := recover()
		if ret, hasReturn := err.(JsValue); hasReturn {
			returnValue = &ret
		} else if err != nil {
			panic(err)
		}
	}()
	s.Scope = NewScope(f.Closure)
	for i, arg := range args {
		if i >= len(f.Params) {
			break
		}
		s.Scope.Set(f.Params[i].Value, &arg)
	}
	for _, stmt := range f.Body.Data.Data {
		EvaluateStmt(stmt, s)
	}
	return returnValue
}

type InterpreterStat struct {
	Program *ast_parser.Stmt
	Scope   *Scope
	Visitor
}

func InitInterpreterStat(program *ast_parser.Stmt, visitor Visitor) *InterpreterStat {
	scope := Scope{
		Parent: nil,
		Env:    map[string]*JsValue{},
	}

	console := NewJsValue(&JsObject{
		Constructor: nil,
		Properties:  map[string]*JsValue{},
	})
	console.Value.(*JsObject).Properties["log"] = NewBuiltIn(func(_ JsValue, args ...JsValue) JsValue {
		for _, arg := range args {
			fmt.Println(ToString(&arg).Value)
		}
		return JsValue{nil, Undefined}
	})
	scope.Set("console", console)

	interpreter := InterpreterStat{
		Program: program,
		Scope:   &scope,
		Visitor: visitor,
	}
	return &interpreter
}

type Scope struct {
	Parent *Scope
	Env    map[string]*JsValue
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent: parent,
		Env:    map[string]*JsValue{},
	}
}

func (s *InterpreterStat) EnterScope() {
	s.Scope = NewScope(s.Scope)
}

func (s *InterpreterStat) PopScope() {
	s.Scope = s.Scope.Parent
}

func (s *Scope) Get(k string) *JsValue {
	if v, has := s.Env[k]; has {
		return v
	} else if s.Parent != nil {
		return s.Parent.Get(k)
	} else {
		return &JsValue{nil, Undefined}
	}
}

func (s *Scope) Set(k string, v *JsValue) {
	if s.Env == nil {
		s.Env = map[string]*JsValue{}
	}
	s.Env[k] = v
}

func (s *Scope) Update(k string, v *JsValue) *JsValue {
	if _, has := s.Env[k]; has {
		s.Env[k] = v
		return v
	} else if s.Parent != nil {
		return s.Parent.Update(k, v)
	} else {
		s.Env[k] = &JsValue{nil, Undefined}
		return &JsValue{nil, Undefined}
	}
}

type RuntimeError struct {
	Loc logger.Loc
	msg string
}

type CodeRunner struct {
	stat *InterpreterStat
	ast  *ast_parser.Stmt
}

func NewCodeRunner(ast *ast_parser.Stmt) CodeRunner {
	return CodeRunner{
		stat: InitInterpreterStat(ast, &IVisitor{}),
		ast:  ast,
	}
}

func (runner *CodeRunner) Run(code string) {
	p := ast_parser.NewParser(code)
	ast := p.Parse()
	EvaluateStmt(&ast, runner.stat)
}

func Run(code string) {
	defer func() {
		err := recover()
		if err != nil {
			if err, isRuntimeErr := err.(RuntimeError); isRuntimeErr {
				log := logger.Logger{Content: code}
				log.Print(logger.LError, err.Loc, err.msg)
			} else {
				panic(err)
			}
		}
	}()

	p := ast_parser.NewParser(code)
	if p.HasError {
		fmt.Println("program stops due to error")
		return
	}
	ast := p.Parse()
	visitor := IVisitor{}
	stat := InitInterpreterStat(&ast, &visitor)

	EvaluateStmt(&ast, stat)
}

func EvaluateExpr(ast *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	switch t := ast.Data.(type) {
	case *ast_parser.EAssign:
		return stat.Visitor.VisitAssignExpr(ast, stat)
	case *ast_parser.ENumericLiteral:
		num, _ := strconv.ParseFloat(t.Value, 64)
		return &JsValue{num, Number}
	case *ast_parser.EStringLiteral:
		return &JsValue{t.Value, String}
	case *ast_parser.EBoolLiteral:
		return &JsValue{t.Value, Boolean}
	case *ast_parser.EBinary:
		return stat.Visitor.VisitBinaryExpr(ast, stat)
	case *ast_parser.EUnary:
		return stat.Visitor.VisitUnaryExpr(ast, stat)
	case *ast_parser.EMemberExpr:
		return stat.Visitor.VisitMemberExpr(ast, stat)
	case *ast_parser.EObjectLiteral:
		return stat.Visitor.VisitObjectExpr(ast, stat)
	case *ast_parser.EArrayLiteral:
		return stat.Visitor.VisitArrayExpr(ast, stat)
	case *ast_parser.ECallExpr:
		return stat.Visitor.VisitCallExpr(ast, stat)
	case *ast_parser.EIdentifier:
		return stat.Scope.Get(t.Value)
	case *ast_parser.EParen:
		return stat.Visitor.VisitParen(ast, stat)
	case *ast_parser.EFunctionExpr:
		return stat.VisitFuncExpr(ast, stat)
	case *ast_parser.EIndex:
		return stat.Visitor.VisitIndexExpr(ast, stat)
	default:
		return &JsValue{nil, Undefined}
	}
}

func EvaluateStmt(ast *ast_parser.Stmt, stat *InterpreterStat) {
	if ast.Data == nil {
		return
	}

	switch ast.Data.(type) {
	case *ast_parser.SProgram:
		stat.Visitor.VisitProgram(ast, stat)
	case *ast_parser.SFor:
		stat.Visitor.VisitForLoopStmt(ast, stat)
	case *ast_parser.SCondition:
		stat.Visitor.VisitConditionStmt(ast, stat)
	case *ast_parser.SFunctionDecl:
		stat.Visitor.VisitFuncStmt(ast, stat)
	case *ast_parser.SExpr:
		stat.Visitor.VisitExpr(&ast.Data.(*ast_parser.SExpr).Expr, stat)
	case *ast_parser.SReturn:
		stat.Visitor.VisitReturnStmt(ast, stat)
	case *ast_parser.SVarDecl:
		stat.Visitor.VisitVarDeclStmt(ast, stat)
	}
}
