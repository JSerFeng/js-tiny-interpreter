package ast_interpreter

import (
	"jsInterpreter/ast_parser"
	"jsInterpreter/logger"
	"strconv"
)

type JsType uint8

const (
	Undefined JsType = iota
	Null
	Number
	String
	Boolean
	Array
	Function
	Object
	BuiltInFunction
	BuiltInObject
)

var JsTypeToString = map[JsType]string{
	Array:    "Array",
	Function: "Function",
	Object:   "Object",
}

type JsValue struct {
	Value interface{}
	Type  JsType
}

func NewJsValue(v interface{}) *JsValue {
	switch v.(type) {
	case string:
		return &JsValue{v, String}
	case int:
		return &JsValue{float64(v.(int)), Number}
	case float32:
		return &JsValue{float64(v.(float32)), Number}
	case float64:
		return &JsValue{v, Number}
	case bool:
		return &JsValue{v, Boolean}
	case *JsFunction:
		return &JsValue{v, Function}
	case *JsObject:
		return &JsValue{v, Object}
	case *JsValue:
		return v.(*JsValue)
	default:
		return &JsValue{nil, Undefined}
	}
}

func NewBuiltIn(builtIn interface{}) *JsValue {
	switch builtIn.(type) {
	case func(JsValue, ...JsValue) JsValue:
		return &JsValue{builtIn, BuiltInFunction}
	case map[string]*JsValue:
		return &JsValue{builtIn, BuiltInObject}
	default:
		panic("unknown builtIn type")
	}
}

func ToString(v *JsValue) JsValue {
	switch v.Type {
	case String:
		return JsValue{v.Value, String}
	case Number:
		return JsValue{strconv.FormatFloat(v.Value.(float64), 'f', 5, 64), String}
	case Boolean:
		if v.Value == true {
			return JsValue{"true", String}
		} else {
			return JsValue{"false", String}
		}
	case Undefined:
		return JsValue{"undefined", String}
	case Null:
		return JsValue{"null", Null}
	default:
		if t, has := JsTypeToString[v.Type]; has {
			return JsValue{"[object " + t + "]", String}
		} else {
			return JsValue{"[object Internal]", String}
		}
	}
}

func ToNumber(v *JsValue) JsValue {
	switch v.Type {
	case Number:
		return JsValue{v.Value, Number}
	case String:
		if value, err := strconv.ParseFloat(v.Value.(string), 64); err != nil {
			return JsValue{0.0, Number}
		} else {
			return JsValue{value, Number}
		}
	case Boolean:
		if v.Value == true {
			return JsValue{1.0, Number}
		} else {
			return JsValue{0.0, Number}
		}
	default:
		return JsValue{0.0, Number}
	}
}

func ToBoolean(v *JsValue) JsValue {
	switch v.Type {
	case Boolean:
		return JsValue{v.Value, Boolean}
	case Number:
		if v.Value == 0 {
			return JsValue{false, Boolean}
		} else {
			return JsValue{true, Boolean}
		}
	case String:
		if v.Value == "" {
			return JsValue{false, Boolean}
		} else {
			return JsValue{true, Boolean}
		}
	default:
		return JsValue{true, Boolean}
	}
}

func IsTruthy(v *JsValue) bool {
	switch v.Type {
	case Boolean:
		return v.Value == true
	case Number:
		return v.Value != 0
	case String:
		return v.Value != ""
	case Undefined, Null:
		return false
	default:
		return true
	}
}

type JsObject struct {
	Constructor *JsValue
	Proto       *JsValue
	Properties  map[string]*JsValue
}

type JsArray struct {
	Arr    []*JsValue
	Length uint
}

type JsFunction struct {
	Closure *Scope
	Params  []ast_parser.EIdentifier
	Body    *ast_parser.SBody
}

var ObjectPrototype = JsValue{map[string]*JsValue{
	"toString": &JsValue{ToString, BuiltInFunction},
}, BuiltInObject}

func ArrayPush(this JsValue, args ...JsValue) JsValue {
	arr := this.Value.(*JsArray)
	items := []*JsValue{}
	for i, _ := range args {
		items = append(items, &args[i])
	}
	arr.Arr = append(arr.Arr, items...)
	arr.Length += uint(len(args))
	return JsValue{arr.Length, Number}
}

func ArraySplice(this JsValue, args ...JsValue) JsValue {
	if len(args) < 2 {
		panic(RuntimeError{logger.Loc{}, "" +
			"the number of the " +
			"arguments of splice " +
			"should be at least 2"})
	}

	index := uint(ToNumber(&args[0]).Value.(float64))
	deleteNum := int(ToNumber(&args[1]).Value.(float64))
	addItems := []*JsValue{}
	for i, _ := range args[2:] {
		addItems = append(addItems, &args[2+i])
	}

	arr := this.Value.(*JsArray)

	if index > arr.Length {
		index = arr.Length
	}

	deleted := []*JsValue{}
	for i := 0; i < deleteNum; i++ {
		deleted = append(deleted, arr.Arr[index])
		arr.Arr = append(arr.Arr[:index], arr.Arr[index+1:]...)
	}

	restArr := arr.Arr[index:]
	arr.Arr = append(arr.Arr[:index], addItems...)
	arr.Arr = append(arr.Arr, restArr...)

	return JsValue{JsArray{deleted, uint(len(deleted))}, Array}
}

var ArrayPrototype = JsValue{map[string]*JsValue{
	"splice": &JsValue{ArraySplice, BuiltInFunction},
	"push":   &JsValue{ArrayPush, BuiltInFunction},
}, BuiltInObject}

func JsToStringFactory(value JsValue) func() JsValue {
	return func() JsValue {
		return ToString(&value)
	}
}

func JsToNumberFactory(value JsValue) func() JsValue {
	return func() JsValue {
		return ToNumber(&value)
	}
}

func JsToBoolFactory(value JsValue) func() JsValue {
	return func() JsValue {
		return ToBoolean(&value)
	}
}

func Wrap(primary *JsValue) JsValue {
	return JsValue{
		Value: &JsValue{map[string]*JsValue{
			"toString": &JsValue{func(this JsValue, args ...JsValue) JsValue {
				return ToString(&this)
			}, BuiltInFunction},
			"toNumber": &JsValue{func(this JsValue, args ...JsValue) JsValue {
				return ToNumber(&this)
			}, BuiltInFunction},
		}, BuiltInObject},
		Type: BuiltInObject,
	}
}
