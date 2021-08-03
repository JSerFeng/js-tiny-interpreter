package lexer

import (
	"jsInterpreter/helper"
	"jsInterpreter/logger"
	"unicode/utf8"
)

type T uint8

const (
	TEndOfFile T = iota
	TLet
	TSyntaxError
	TSingleLineComment
	TMultiLineComment
	// "#!/usr/bin/env node"
	THashbang

	// Literals
	TNoSubstitutionTemplateLiteral // Contents are in lexer.StringLiteral ([]uint16)
	TNumericLiteral                // Contents are in lexer.Number (float64)
	TStringLiteral                 // Contents are in lexer.StringLiteral ([]uint16)
	TBigIntegerLiteral             // Contents are in lexer.Identifier (string)

	// Pseudo-literals
	TTemplateHead   // Contents are in lexer.StringLiteral ([]uint16)
	TTemplateMiddle // Contents are in lexer.StringLiteral ([]uint16)
	TTemplateTail   // Contents are in lexer.StringLiteral ([]uint16)

	// Punctuation
	TAmpersand
	TAmpersandAmpersand
	TAsterisk
	TAsteriskAsterisk
	TAt
	TBar
	TBarBar
	TCaret //^
	TCloseBrace
	TCloseBracket
	TCloseParen
	TColon
	TComma
	TDot
	TDotDotDot
	TEqualsEquals
	TEqualsEqualsEquals
	TEqualsGreaterThan
	TExclamation
	TExclamationEquals
	TExclamationEqualsEquals
	TGreaterThan
	TGreaterThanEquals
	TGreaterThanGreaterThan
	TGreaterThanGreaterThanGreaterThan
	TLessThan
	TLessThanEquals
	TLessThanLessThan
	TMinus
	TMinusMinus
	TOpenBrace
	TOpenBracket
	TOpenParen
	TPercent
	TPlus
	TPlusPlus
	TQuestion
	TQuestionDot
	TQuestionQuestion
	TSemicolon
	TSlash
	TTilde

	// Assignments (keep in sync with IsAssign() below)
	TAmpersandAmpersandEquals
	TAmpersandEquals
	TAsteriskAsteriskEquals
	TAsteriskEquals
	TBarBarEquals
	TBarEquals
	TCaretEquals
	TEquals
	TGreaterThanGreaterThanEquals
	TGreaterThanGreaterThanGreaterThanEquals
	TLessThanLessThanEquals
	TMinusEquals
	TPercentEquals
	TPlusEquals
	TQuestionQuestionEquals
	TSlashEquals

	// Class-private fields and methods
	TPrivateIdentifier

	// Identifiers
	TIdentifier     // Contents are in lexer.Identifier (string)
	TEscapedKeyword // A keyword that has been escaped as an identifer

	// Reserved words
	TBreak
	TCase
	TCatch
	TClass
	TConst
	TContinue
	TDebugger
	TDefault
	TDelete
	TDo
	TElse
	TEnum
	TExport
	TExtends
	TFalse
	TFinally
	TFor
	TFunction
	TIf
	TImport
	TIn
	TInstanceof
	TNew
	TNull
	TReturn
	TSuper
	TSwitch
	TThis
	TThrow
	TTrue
	TTry
	TTypeof
	TVar
	TVoid
	TWhile
	TWith
)

var Keywords = map[string]T{
	// Reserved words
	"let":        TLet,
	"break":      TBreak,
	"case":       TCase,
	"catch":      TCatch,
	"class":      TClass,
	"const":      TConst,
	"continue":   TContinue,
	"debugger":   TDebugger,
	"default":    TDefault,
	"delete":     TDelete,
	"do":         TDo,
	"else":       TElse,
	"enum":       TEnum,
	"export":     TExport,
	"extends":    TExtends,
	"finally":    TFinally,
	"for":        TFor,
	"function":   TFunction,
	"if":         TIf,
	"import":     TImport,
	"in":         TIn,
	"instanceof": TInstanceof,
	"new":        TNew,
	"null":       TNull,
	"return":     TReturn,
	"super":      TSuper,
	"switch":     TSwitch,
	"this":       TThis,
	"throw":      TThrow,
	"try":        TTry,
	"typeof":     TTypeof,
	"var":        TVar,
	"void":       TVoid,
	"while":      TWhile,
	"with":       TWith,
}

var Token2StringMap = map[T]string{
	// "#!/usr/bin/env node"
	THashbang: "#!",

	// Punctuation
	TAmpersand:                         "&",
	TAmpersandAmpersand:                "&&",
	TAsterisk:                          "*",
	TAsteriskAsterisk:                  "**",
	TAt:                                "@",
	TBar:                               "|",
	TBarBar:                            "||",
	TCaret:                             "^",
	TCloseBrace:                        "}",
	TCloseBracket:                      "]",
	TCloseParen:                        ")",
	TColon:                             ":",
	TComma:                             ",",
	TDot:                               ".",
	TDotDotDot:                         "...",
	TEqualsEquals:                      "==",
	TEqualsEqualsEquals:                "===",
	TEqualsGreaterThan:                 ">=",
	TExclamation:                       "!",
	TExclamationEquals:                 "!=",
	TExclamationEqualsEquals:           "!==",
	TGreaterThan:                       ">",
	TGreaterThanEquals:                 ">=",
	TGreaterThanGreaterThan:            ">>",
	TGreaterThanGreaterThanGreaterThan: ">>>",
	TLessThan:                          "<",
	TLessThanEquals:                    "<=",
	TLessThanLessThan:                  "<<",
	TMinus:                             "-",
	TMinusMinus:                        "--",
	TOpenBrace:                         "{",
	TOpenBracket:                       "[",
	TOpenParen:                         "(",
	TPercent:                           "%",
	TPlus:                              "+",
	TPlusPlus:                          "++",
	TQuestion:                          "?",
	TQuestionDot:                       "?.",
	TQuestionQuestion:                  "??",
	TSemicolon:                         ";",
	TSlash:                             "/",
	TTilde:                             "~",

	// Assignments (keep in sync with IsAssign() below)
	TAmpersandAmpersandEquals:                "&&=",
	TAmpersandEquals:                         "&=",
	TAsteriskAsteriskEquals:                  "**=",
	TAsteriskEquals:                          "*=",
	TBarBarEquals:                            "||=",
	TBarEquals:                               "|=",
	TCaretEquals:                             "^=",
	TEquals:                                  "=",
	TGreaterThanGreaterThanEquals:            ">>=",
	TGreaterThanGreaterThanGreaterThanEquals: ">>>=",
	TLessThanLessThanEquals:                  "<<=",
	TMinusEquals:                             "-=",
	TPercentEquals:                           "%=",
	TPlusEquals:                              "+=",
	TQuestionQuestionEquals:                  "??=",
	TSlashEquals:                             "/=",

	// Class-private fields and methods
	TPrivateIdentifier: "#",

	TLet:        "let",
	TBreak:      "break",
	TCase:       "case",
	TCatch:      "catch",
	TClass:      "class",
	TConst:      "const",
	TContinue:   "continue",
	TDebugger:   "debugger",
	TDefault:    "default",
	TDelete:     "delete",
	TDo:         "do",
	TElse:       "else",
	TEnum:       "enum",
	TExport:     "export",
	TExtends:    "extends",
	TFalse:      "false",
	TFinally:    "finally",
	TFor:        "for",
	TFunction:   "function",
	TIf:         "if",
	TImport:     "import",
	TIn:         "in",
	TInstanceof: "instanceof",
	TNew:        "new",
	TNull:       "null",
	TReturn:     "return",
	TSuper:      "super",
	TSwitch:     "switch",
	TThis:       "this",
	TThrow:      "throw",
	TTrue:       "true",
	TTry:        "try",
	TTypeof:     "typeof",
	TVar:        "var",
	TVoid:       "void",
	TWhile:      "while",
	TWith:       "with",
}

type Token struct {
	T   T
	Loc logger.Loc
}

type Lexer struct {
	Log          logger.Logger
	Current      int
	End          int
	Line         int
	Source       string
	RawSource    string
	Identifier   string
	CurChar      rune
	CurCharWidth int
	HasError     bool
}

func (lexer *Lexer) IsEnd() bool {
	return lexer.Current >= len(lexer.RawSource)
}

func (lexer *Lexer) Error(loc logger.Loc, msg string) {
	lexer.HasError = true
	lexer.Log.Print(logger.LError, loc, msg)
}

func (lexer *Lexer) Raw(loc logger.Loc) string {
	return lexer.RawSource[loc.Offset : loc.Len+loc.Offset]
}

func (lexer *Lexer) Tokenizer() []Token {
	tokens := make([]Token, 0)
	defer func() {
		if err := recover(); err != nil {
			lexer.HasError = true
			lexer.Tokenizer()
		}
	}()
	for !lexer.IsEnd() {
		var token Token
		loc := logger.Loc{
			Offset: lexer.Current,
			Len:    lexer.CurCharWidth,
			Line:   lexer.Line,
		}
		curChar := lexer.CurChar
		lexer.Step()
		switch curChar {
		case ' ', '\t':
			continue
		case '\n', '\r':
			lexer.Line++
			continue
		case ':':
			token = Token{TColon, loc}
		case ';':
			token = Token{TSemicolon, loc}
		case ',':
			token = Token{TComma, loc}
		case '.':
			if helper.IsValidNumber(string(lexer.CurChar)) {
			DotNumeric:
				for {
					switch lexer.CurChar {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						loc.Len += lexer.CurCharWidth
						lexer.Step()
					default:
						break DotNumeric
					}
				}
				token = Token{TNumericLiteral, loc}
			} else {
				token = Token{TDot, loc}
			}
		case '(':
			token = Token{TOpenParen, loc}
		case ')':
			token = Token{TCloseParen, loc}
		case '{':
			token = Token{TOpenBrace, loc}
		case '}':
			token = Token{TCloseBrace, loc}
		case '[':
			token = Token{TOpenBracket, loc}
		case ']':
			token = Token{TCloseBracket, loc}
		case '<':
			switch lexer.CurChar {
			case '=':
				token = Token{TLessThanEquals, loc}
			default:
				token = Token{TLessThan, loc}
			}
		case '>':
			switch lexer.CurChar {
			case '=':
				token = Token{TGreaterThanEquals, loc}
			default:
				token = Token{TGreaterThan, loc}
			}
		case '!':
			switch lexer.CurChar {
			case '=':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				if lexer.CurChar == '=' {
					loc.Len += lexer.CurCharWidth
					lexer.Step()
					token = Token{TExclamationEqualsEquals, loc}
				} else {
					token = Token{TExclamationEquals, loc}
				}
			default:
				token = Token{TExclamation, loc}
			}
		case '=':
			switch lexer.CurChar {
			case '=':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				if lexer.CurChar == '=' {
					loc.Len = lexer.CurCharWidth
					lexer.Step()
					token = Token{TEqualsEqualsEquals, loc}
				}
				token = Token{TEqualsEquals, loc}
			case '>':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TEqualsGreaterThan, loc}
			default:
				token = Token{TEquals, loc}
			}
		case '+':
			switch lexer.CurChar {
			case '=':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TPlusEquals, loc}
			case '+':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TPlusPlus, loc}
			default:
				token = Token{TPlus, loc}
			}
		case '-':
			switch lexer.CurChar {
			case '=':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TMinusEquals, loc}
			case '-':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TMinusMinus, loc}
			case '>':
				token = Token{TEqualsGreaterThan, loc}
			default:
				token = Token{TMinus, loc}
			}
		case '*':
			switch lexer.CurChar {
			case '*':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				if lexer.CurChar == '=' {
					loc.Len += lexer.CurCharWidth
					lexer.Step()
					token = Token{TAsteriskAsteriskEquals, loc}
				} else {
					token = Token{TAsteriskAsterisk, loc}
				}
			case '=':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TAsteriskEquals, loc}
			default:
				token = Token{TAsterisk, loc}
			}
		case '/':
			switch lexer.CurChar {
			case '=':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				token = Token{TSlashEquals, loc}
			case '/':
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				for lexer.CurChar != '\n' {
					loc.Len += lexer.CurCharWidth
					lexer.Step()
				}
				token = Token{TSingleLineComment, loc}
			case '*':
				//多行注释/** abcd */
				loc.Len += lexer.CurCharWidth
				lexer.Step()
				for lexer.CurChar != '/' {
					if lexer.CurChar == '\\' {
						loc.Len += lexer.CurCharWidth
						lexer.Step()
					}
					loc.Len += lexer.CurCharWidth
					lexer.Step()
				}
			default:
				token = Token{TSlash, loc}
			}
		case '%':
			token = Token{TPercent, loc}
		case '_', '$',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		Identifier:
			for !lexer.IsEnd() {
				switch lexer.CurChar {
				case '_', '$',
					'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
					'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
					'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
					'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
					'0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					loc.Len += lexer.CurCharWidth
					lexer.Step()
				default:
					break Identifier
				}
			}
			name := lexer.Raw(loc)
			if name == "true" {
				token = Token{TTrue, loc}
			} else if name == "false" {
				token = Token{TFalse, loc}
			} else if t, ok := Keywords[name]; ok {
				token = Token{t, loc}
			} else {
				token = Token{TIdentifier, loc}
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			num := string(curChar)
		Numeric:
			for !lexer.IsEnd() {
				switch lexer.CurChar {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
					num += string(lexer.CurChar)
					loc.Len += lexer.CurCharWidth
					lexer.Step()
				default:
					break Numeric
				}
			}
			if !helper.IsValidNumber(num) {
				lexer.Error(loc, "unknown numeric literal")
			}
			token = Token{TNumericLiteral, loc}
		case '"', '\'':
			//String Literal
			t := curChar
			for !lexer.IsEnd() && lexer.CurChar != t {
				loc.Len += lexer.CurCharWidth
				lexer.Step()
			}
			loc.Len += lexer.CurCharWidth
			if !lexer.IsEnd() {
				lexer.Step()
			}
			token = Token{TStringLiteral, loc}
		default:
			lexer.Error(loc, "unexpected word")
		}
		tokens = append(tokens, token)
	}

	return tokens
}

func Raw(loc logger.Loc, content string) string {
	return content[loc.Offset : loc.Offset+loc.Len]
}

func (lexer *Lexer) Step() {
	if lexer.IsEnd() {
		return
	}
	char, width := utf8.DecodeRuneInString(lexer.Source)

	lexer.CurChar = char
	lexer.CurCharWidth = width
	lexer.Current = lexer.End
	lexer.End += width
	lexer.Source = lexer.Source[width:]
}

func NewLexer(rawSource string) Lexer {
	lexer := Lexer{
		Log: logger.Logger{
			Content: rawSource,
		},
		Current:   0,
		End:       0,
		Source:    rawSource,
		RawSource: rawSource,
		Line:      1,
		HasError:  false,
	}
	lexer.Step()
	return lexer
}
