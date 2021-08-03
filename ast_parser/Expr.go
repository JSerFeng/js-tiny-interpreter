package ast_parser

import (
	"jsInterpreter/lexer"
	"jsInterpreter/logger"
)

type E interface{ IsExpr() }

type Expr struct {
	Loc  logger.Loc
	Data E
}

type EOp lexer.T

type EAssign struct {
	Target     *Expr
	Assignment *Expr
}

type EBinary struct {
	Op    EOp
	Left  Expr
	Right Expr
}

//左结合还是右结合
type Association uint8

const (
	AssociationLeft Association = iota
	AssociationRight
)

type EUnary struct {
	Op    EOp
	Value Expr
	Association
}

type EIndex struct {
	Target Expr
	Idx    Expr
}

/*
a.b.c对应
Obj: EMemberExpr{
	Obj: 		a,
	Property: b
}
Property: c
*/
type EMemberExpr struct {
	Obj      Expr
	Property EIdentifier
}

type EParen struct {
	Data Expr
}

type EStringLiteral struct {
	Value string
}

type ENumericLiteral struct {
	Value string
}

type EIdentifier struct {
	Value string
}

type EArrowFunction struct {
	Id     EIdentifier
	Params []EIdentifier
	Body   SBlock
}

type EFunctionExpr struct {
	Id     *EIdentifier
	Params []EIdentifier
	Body   *SBody
}

type EBoolLiteral struct {
	Value bool
}

type ECallExpr struct {
	Callee Expr
	Args   []*Expr
}

type EArrayLiteral struct {
	Arr    []*Expr
	Length uint
}

type EObjectLiteral struct {
	Proto       *EIdentifier
	Constructor *EIdentifier
	Properties  []*Expr
}

func (e *EOp) IsExpr()             {}
func (e *EBinary) IsExpr()         {}
func (e *EParen) IsExpr()          {}
func (e *EUnary) IsExpr()          {}
func (e *EAssign) IsExpr()         {}
func (e *EStringLiteral) IsExpr()  {}
func (e *ENumericLiteral) IsExpr() {}
func (e *EIdentifier) IsExpr()     {}
func (e *EArrowFunction) IsExpr()  {}
func (e *EFunctionExpr) IsExpr()   {}
func (e *EBoolLiteral) IsExpr()    {}
func (e *EMemberExpr) IsExpr()     {}
func (e *ECallExpr) IsExpr()       {}
func (e *EArrayLiteral) IsExpr()   {}
func (e *EObjectLiteral) IsExpr()  {}
func (e *EIndex) IsExpr()          {}
