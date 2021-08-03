package main

import (
	"fmt"
	ast_interpreter "jsInterpreter/ast-interpreter"
	"jsInterpreter/cmd"
	"os"
)

func main() {

	for _, item := range jsArr.Value.(*ast_interpreter.JsArray).Arr {
		fmt.Println(item.Value)
	}

	args := os.Args
	switch len(args) {
	case 1: // 命令行
		cmd.RunCommand()
	case 2: // 输入的文件地址
		dir := args[1]
		if data, err := os.ReadFile(dir); err != nil {
			fmt.Println("找不到文件")
		} else {
			ast_interpreter.Run(string(data))
		}

	default:
		fmt.Println("参数错误\n" +
			"输入一个目标代码相对位置或直接以命令行执行")
	}
}
