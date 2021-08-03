package ast_interpreter

import (
	"fmt"
	t "testing"
)

func TestArrayMethods(t *t.T) {
	jsArr := JsValue{&JsArray{
		Arr: []*JsValue{
			&JsValue{0.0, Number},
			&JsValue{1.0, Number},
			&JsValue{2.0, Number},
		},
		Length: 1,
	}, Array}
	arr := []float64{.0, 1.0, 2.0}
	array := jsArr.Value.(*JsArray)

	ArrayPush(jsArr, JsValue{3.0, Number}, JsValue{4.0, Number})
	arr = append(arr, 3.0, 4.0)

	checkEqual(t, array, arr)


	ArraySplice(jsArr, JsValue{1.0, Number}, JsValue{1.0, Number})
	arr = append(arr[:1], arr[2:]...)

	checkEqual(t, array, arr)
}

func checkEqual(t *t.T, array *JsArray, arr []float64) {
	if array.Length != uint(len(arr)) {
		t.Error("长度应该为" + fmt.Sprint(len(arr)) + "，结果为" + fmt.Sprint(array.Length))
	}

	for i, item := range arr {
		if array.Arr[i].Value != item {
			t.Error("数组中元素应该和原生一致，")
		}
	}
}
