package ast_parser

import "jsInterpreter/logger"

type S interface{ IsStmt() }

type Stmt struct {
	Loc  logger.Loc
	Data S
}

type VarKind uint8

const (
	VLet VarKind = iota
	VConst
	VVar
)

type SVarDecl struct {
	Id   EIdentifier
	kind VarKind
	Init *Expr
}

type SExpr struct{ Expr }

type ClassMemberKind uint8

const (
	PropertyDefinition ClassMemberKind = iota
	MethodDefinition
)

type SProgram struct {
	Body []*Stmt
}

type SBlock struct {
	Data []*Stmt
}

type SBreak struct{}
type SContinue struct{}

type SBody struct {
	Loc  logger.Loc
	Data *SBlock
}

//for loop
type SFor struct {
	Initializer *Stmt
	Condition   *Expr
	Reset       *Stmt
	Body        *SBody
}

//while loop
type SWhile struct {
	Condition Stmt
	Body      SBody
}

//if-else statement
// if condition body else if condition body else body
type SCondition struct{ Branches []Branch }

type Branch struct {
	Condition *Expr
	Body      *SBody
}

//function
type SFunctionDecl struct {
	Id     *EIdentifier
	Params []EIdentifier
	Body   *SBody
}

type SReturn struct {
	Expr
}

//class
type SClass struct {
	SuperClass *Expr
	Id         EIdentifier
	Body       []ClassMember
}

type ClassMember struct {
	Kind  ClassMemberKind
	Loc   logger.Loc
	Key   EIdentifier
	Value ClassMemberValue
}

type ClassMemberValue interface{ IsClassMemberValue() }

func (s *SFunctionDecl) IsClassMemberValue() {}
func (p *SVarDecl) IsClassMemberValue()      {}

func (s *SVarDecl) IsStmt()      {}
func (s *SExpr) IsStmt()         {}
func (s *SClass) IsStmt()        {}
func (s *SFunctionDecl) IsStmt() {}
func (s *SBlock) IsStmt()        {}
func (s *SProgram) IsStmt()      {}
func (s *SFor) IsStmt()          {}
func (s *SWhile) IsStmt()        {}
func (s *SBody) IsStmt()         {}
func (s *SCondition) IsStmt()    {}
func (s *SReturn) IsStmt()       {}
func (s *SBreak) IsStmt()        {}
func (s *SContinue) IsStmt()     {}
