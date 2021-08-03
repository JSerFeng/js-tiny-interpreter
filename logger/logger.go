package logger

import (
	"fmt"
	"unicode/utf8"
)

type Logger struct {
	Content string
}

type Loc struct {
	Offset int
	Len    int
	Line   int
}

type LKind uint

const (
	LError LKind = iota
	LWarn
)

func (log *Logger) Print(kind LKind, loc Loc, errMsg string) {
	//计算多少行
	content := log.Content
	line := 1
	start := 0
	for {
		if line == loc.Line {
			break
		}
		if len(content) == 0 {
			break
		}
		c, w := utf8.DecodeRuneInString(content)
		if c == '\n' {
			line++
		}
		start += w
		content = content[w:]
	}

	end := start
	for {
		if len(content) == 0 {
			break
		}
		c, w := utf8.DecodeRuneInString(content)
		if c == '\n' {
			break
		}
		end += w
		content = content[w:]
	}
	//start - end是这一行的文本
	rawCode := log.Content[start:end]
	fmt.Printf("%6d| %s\n", line, rawCode)
	fmt.Printf("%8s", "")
	for i := 0; i < loc.Offset-start; i++ {
		if rawCode[i] == '\t' {
			fmt.Printf("\t")
		} else {
			fmt.Printf(" ")
		}
	}
	for i := 0; i < loc.Len; i++ {
		fmt.Printf("^")
	}
	fmt.Println("\n", errMsg)
}
