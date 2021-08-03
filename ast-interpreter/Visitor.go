package ast_interpreter

import (
	"jsInterpreter/ast_parser"
	"jsInterpreter/lexer"
	"strconv"
)

type Visitor interface {
	//statement
	VisitProgram(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitVarDeclStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitFuncStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitReturnStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitConditionStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitForLoopStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitBreakStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitContinueStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitBlockStmt(s *ast_parser.Stmt, stat *InterpreterStat)
	VisitStmts(stmts []*ast_parser.Stmt, stat *InterpreterStat)

	//expression
	VisitExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitFuncExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitCallExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitNumericLiteral(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitStringLiteral(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitBoolLiteral(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitBinaryExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitUnaryExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitMemberExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitParen(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitFunctionExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitAssignExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitArrayExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitObjectExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
	VisitIndexExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue
}

type IVisitor struct{}

func (v *IVisitor) VisitProgram(s *ast_parser.Stmt, stat *InterpreterStat) {
	for _, stmt := range s.Data.(*ast_parser.SProgram).Body {
		EvaluateStmt(stmt, stat)
	}
}

func (v *IVisitor) VisitBlockStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	for _, stmt := range s.Data.(*ast_parser.SBlock).Data {
		EvaluateStmt(stmt, stat)
	}
}

func (v *IVisitor) VisitForLoopStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	stat.EnterScope()
	defer func() {
		stat.PopScope()
		err := recover()
		if err != nil {
			switch err.(type) {
			case ast_parser.SBreak:
			case ast_parser.SContinue:
			default:
				panic(err)
			}
		}
	}()
	forLoop := s.Data.(*ast_parser.SFor)
	EvaluateStmt(forLoop.Initializer, stat)
	runLoop(forLoop, stat)
}

func runLoop(forLoop *ast_parser.SFor, stat *InterpreterStat) {
	defer func() {
		err := recover()
		switch err.(type) {
		case ast_parser.SContinue:
			EvaluateStmt(forLoop.Reset, stat)
			runLoop(forLoop, stat)
		case ast_parser.SBreak: //do nothing
		case nil: //do nothing
		default:
			panic(err)
		}
	}()
	for IsTruthy(EvaluateExpr(forLoop.Condition, stat)) {
		for _, stmt := range forLoop.Body.Data.Data {
			EvaluateStmt(stmt, stat)
		}
		EvaluateStmt(forLoop.Reset, stat)
	}
}

func (v *IVisitor) VisitBreakStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	panic(ast_parser.SBreak{})
}
func (v *IVisitor) VisitContinueStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	panic(ast_parser.SContinue{})
}

func (v *IVisitor) VisitConditionStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	stat.EnterScope()
	defer func() {
		stat.PopScope()
	}()
	conditionStmt := s.Data.(*ast_parser.SCondition)
	for _, branch := range conditionStmt.Branches {
		if IsTruthy(EvaluateExpr(branch.Condition, stat)) {
			v.VisitStmts(branch.Body.Data.Data, stat)
			break
		}
	}
}

func (v *IVisitor) VisitFuncStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	fnDecl := s.Data.(*ast_parser.SFunctionDecl)
	fn := NewJsValue(JsFunction{
		Closure: stat.Scope,
		Params:  fnDecl.Params,
		Body:    fnDecl.Body,
	})

	if fnDecl.Id != nil {
		stat.Scope.Set(fnDecl.Id.Value, fn)
	}
}

func (v *IVisitor) VisitReturnStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	panic(EvaluateExpr(&s.Data.(*ast_parser.SReturn).Expr, stat))
}

func (v *IVisitor) VisitStmts(stmts []*ast_parser.Stmt, stat *InterpreterStat) {
	for _, stmt := range stmts {
		EvaluateStmt(stmt, stat)
	}
}

func (v *IVisitor) VisitVarDeclStmt(s *ast_parser.Stmt, stat *InterpreterStat) {
	varDecl := s.Data.(*ast_parser.SVarDecl)
	if varDecl.Init != nil {
		stat.Scope.Set(varDecl.Id.Value, EvaluateExpr(varDecl.Init, stat))
	} else {
		stat.Scope.Set(varDecl.Id.Value, &JsValue{nil, Undefined})
	}
}

/****************/
/** expression **/
/****************/
func (v *IVisitor) VisitExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	return EvaluateExpr(e, stat)
}

func (v *IVisitor) VisitBinaryExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	binaryExpr := e.Data.(*ast_parser.EBinary)
	op := binaryExpr.Op

	lValue := EvaluateExpr(&binaryExpr.Left, stat)
	rValue := EvaluateExpr(&binaryExpr.Right, stat)

	switch op {
	case
		ast_parser.EOp(lexer.TMinus),
		ast_parser.EOp(lexer.TSlash),
		ast_parser.EOp(lexer.TAsterisk),
		ast_parser.EOp(lexer.TPercent):
		lNum := ToNumber(lValue).Value.(float64)
		rNum := ToNumber(rValue).Value.(float64)

		switch op {
		case
			ast_parser.EOp(lexer.TMinus):
			return &JsValue{lNum - rNum, Number}
		case ast_parser.EOp(lexer.TSlash):
			return &JsValue{lNum / rNum, Number}
		case ast_parser.EOp(lexer.TAsterisk):
			return &JsValue{lNum * rNum, Number}
		case ast_parser.EOp(lexer.TPercent):
			return &JsValue{int(lNum) % int(rNum), Number}
		default:
			panic(RuntimeError{e.Loc, "operand not allowed"})
		}
	case ast_parser.EOp(lexer.TPlus):
		isStringPlus := false
		//判断是否是字符串相加
		switch lValue.Type {
		case String:
			isStringPlus = true
			goto Calc
		}
		switch rValue.Type {
		case String:
			isStringPlus = true
			goto Calc
		}

	Calc:
		if isStringPlus {
			return &JsValue{ToString(lValue).Value.(string) + ToString(rValue).Value.(string), String}
		} else {
			lNum := ToNumber(lValue)
			return &JsValue{lNum.Value.(float64) + ToNumber(rValue).Value.(float64), Number}
		}
	case ast_parser.EOp(lexer.TEqualsEquals):
		return &JsValue{
			Value: lValue.Type == rValue.Type && lValue.Value == rValue.Value,
			Type:  Boolean,
		}
	case ast_parser.EOp(lexer.TLessThan),
		ast_parser.EOp(lexer.TGreaterThan),
		ast_parser.EOp(lexer.TLessThanEquals),
		ast_parser.EOp(lexer.TGreaterThanEquals):
		if lValue.Type != rValue.Type {
			panic(RuntimeError{e.Loc, "left type mismatch right type"})
		}
		if lValue.Type != Number {
			panic(RuntimeError{e.Loc, "type must be number"})
		}
		switch op {
		case ast_parser.EOp(lexer.TLessThan):
			return &JsValue{lValue.Value.(float64) < rValue.Value.(float64), Boolean}
		case ast_parser.EOp(lexer.TGreaterThan):
			return &JsValue{lValue.Value.(float64) > rValue.Value.(float64), Boolean}
		case ast_parser.EOp(lexer.TLessThanEquals):
			return &JsValue{lValue.Value.(float64) <= rValue.Value.(float64), Boolean}
		case ast_parser.EOp(lexer.TGreaterThanEquals):
			return &JsValue{lValue.Value.(float64) >= rValue.Value.(float64), Boolean}
		default:
			return &JsValue{false, Boolean}
		}
	default:
		panic(RuntimeError{e.Loc, "unknown operator"})
	}
}

func (v *IVisitor) VisitUnaryExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	unary := e.Data.(*ast_parser.EUnary)

	val := EvaluateExpr(&unary.Value, stat)

	switch unary.Op {
	case ast_parser.EOp(lexer.TExclamation):
		return &JsValue{!ToBoolean(val).Value.(bool), Boolean}
	case
		ast_parser.EOp(lexer.TPlusPlus),
		ast_parser.EOp(lexer.TMinusMinus):

		if val.Type != Number {
			panic(RuntimeError{e.Loc, "this operator need number operand"})
		}
		target := 1.0
		if unary.Op == ast_parser.EOp(lexer.TMinusMinus) {
			target = -1.0
		}

		nextVal := &JsValue{
			Value: val.Value.(float64) + target,
			Type:  Number,
		}

		switch t := unary.Value.Data.(type) {
		case *ast_parser.EIdentifier:
			key := t.Value
			if unary.Association == ast_parser.AssociationRight {
				defer func() {
					stat.Scope.Update(key, nextVal)
				}()
				return val
			}

			return stat.Scope.Update(key, nextVal)
		case *ast_parser.EMemberExpr:
			val.Value = nextVal.Value
			return nextVal
		default:
			panic(RuntimeError{e.Loc, "unknown type"})
		}

	default:
		panic(RuntimeError{e.Loc, "unknown unary operator"})
	}
}

func (v *IVisitor) VisitFuncExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	fnDecl := e.Data.(*ast_parser.EFunctionExpr)
	fn := &JsValue{&JsFunction{
		Closure: stat.Scope,
		Params:  fnDecl.Params,
		Body:    fnDecl.Body,
	}, Function}
	if fnDecl.Id != nil {
		stat.Scope.Set(fnDecl.Id.Value, fn)
	}
	return fn
}

func (v *IVisitor) VisitCallExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	callExpr := e.Data.(*ast_parser.ECallExpr)
	calleeValue := EvaluateExpr(&callExpr.Callee, stat)
	if calleeValue.Type != Function && calleeValue.Type != BuiltInFunction {
		panic(RuntimeError{e.Loc, " is not a function"})
	}

	//找到"this"
	var _this JsValue = JsValue{nil, Undefined}
	switch callee := callExpr.Callee.Data.(type) {
	case *ast_parser.EMemberExpr:
		//这里对于基本类型进行装箱
		_this = *EvaluateExpr(&callee.Obj, stat)
	}

	args := []JsValue{}
	for _, arg := range callExpr.Args {
		_arg := EvaluateExpr(arg, stat)
		args = append(args, *_arg)
	}

	if calleeValue.Type == BuiltInFunction {
		result := calleeValue.Value.(func(JsValue, ...JsValue) JsValue)(_this, args...)
		return &result
	} else {
		callee := calleeValue.Value.(*JsFunction)
		return stat.Call(callee, args)
	}
}

func (v *IVisitor) VisitFunctionExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	fnExpr := e.Data.(*ast_parser.EFunctionExpr)
	return &JsValue{&JsFunction{
		Closure: stat.Scope,
		Params:  fnExpr.Params,
		Body:    fnExpr.Body,
	}, Function}
}

func (v *IVisitor) VisitMemberExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	memberExpr := e.Data.(*ast_parser.EMemberExpr)
	obj := EvaluateExpr(&memberExpr.Obj, stat)
	property := memberExpr.Property.Value
	switch obj.Type {
	case Object:
		val := obj.Value.(*JsObject).Properties[property]
		return val
	case BuiltInObject:
		return obj.Value.(map[string]*JsValue)[property]
	default:
		//装箱
		valWrap := Wrap(obj)
		return valWrap.Value.(*JsValue).Value.(map[string]*JsValue)[property]
	}
}

func (v *IVisitor) VisitNumericLiteral(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	if num, err := strconv.ParseFloat(e.Data.(*ast_parser.ENumericLiteral).Value, 64); err == nil {
		return &JsValue{
			num,
			Number}
	} else {
		panic(RuntimeError{e.Loc, "not a number"})
	}
}

func (v *IVisitor) VisitParen(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	return EvaluateExpr(&e.Data.(*ast_parser.EParen).Data, stat)
}

func (v *IVisitor) VisitStringLiteral(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	return &JsValue{e.Data.(*ast_parser.ENumericLiteral).Value, String}
}

func (v *IVisitor) VisitBoolLiteral(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	return &JsValue{e.Data.(*ast_parser.ENumericLiteral).Value == "true", Boolean}
}

func (v *IVisitor) VisitAssignExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	assignExpr := e.Data.(*ast_parser.EAssign)
	assignTarge := EvaluateExpr(assignExpr.Target, stat)

	assignValue := EvaluateExpr(assignExpr.Assignment, stat)
	*assignTarge = *assignValue
	return assignTarge
}

func (v *IVisitor) VisitArrayExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	arrLiteral := e.Data.(*ast_parser.EArrayLiteral)
	arr := []*JsValue{}

	for _, expr := range arrLiteral.Arr {
		arr = append(arr, EvaluateExpr(expr, stat))
	}

	return &JsValue{
		Value: &JsArray{
			Arr:    arr,
			Length: arrLiteral.Length,
		},
		Type: Array,
	}
}

func (v *IVisitor) VisitObjectExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	objExpr := e.Data.(*ast_parser.EObjectLiteral)
	properties := map[string]*JsValue{}
	for i := 0; i < len(objExpr.Properties); i += 2 {
		k := EvaluateExpr(objExpr.Properties[i], stat)
		val := EvaluateExpr(objExpr.Properties[i+1], stat)
		if k.Type == String {
			properties[k.Value.(string)] = val
		} else if k.Type == Number {
			str := strconv.FormatFloat(k.Value.(float64), 'f', 5, 64)
			properties[str] = val
		} else {
			panic(RuntimeError{e.Loc, JsTypeToString[k.Type] + " is not allowed to be keys"})
		}
	}

	var proto *JsValue = nil
	if objExpr.Proto != nil {
		proto = stat.Scope.Get(objExpr.Proto.Value)
	} else {
		proto = &ObjectPrototype
	}

	return &JsValue{
		Value: &JsObject{
			Constructor: nil,
			Proto:       proto,
			Properties:  properties,
		},
		Type: Object,
	}
}

func (v *IVisitor) VisitIndexExpr(e *ast_parser.Expr, stat *InterpreterStat) *JsValue {
	idxExpr := e.Data.(*ast_parser.EIndex)
	targetArr := EvaluateExpr(&idxExpr.Target, stat)
	idx := EvaluateExpr(&idxExpr.Idx, stat)
	switch targetArr.Type {
	case Array:
		arr := targetArr.Value.(*JsArray)
		i := uint(ToNumber(idx).Value.(float64))
		if i >= arr.Length {
			return &JsValue{nil, Undefined}
		}
		return arr.Arr[i]
	default:
		panic(RuntimeError{e.Loc, "index expression can only be used for array or object"})
	}
}
