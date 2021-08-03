package cmd

import (
	"bufio"
	"fmt"
	ast_interpreter "jsInterpreter/ast-interpreter"
	"os"
	"strings"
)

func RunCommand() {
	var code string
	var prev string
	fmt.Println("TIP: 如果想要换行请输入\\n后点击回车")
	runner := ast_interpreter.NewCodeRunner(nil)
	for {
		print("> ")
		reader := bufio.NewReader(os.Stdin)
		_code, err := reader.ReadString('\n')
		code = _code
		if err !=nil {
			continue
		}
		if strings.HasSuffix(code, "\n\n") {
			//允许用户换行输入而不是直接执行
			prev += code
			continue
		} else {
			code += prev
			runner.Run(code)
			prev = ""
		}
	}
}
