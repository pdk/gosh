package token

import "strconv"

// Token is what particular type of token found.
type Token int

// The list of tokens
const (
	ILLEGAL             Token = iota
	NADA                      // no token
	EOF                       // END OF FILE
	COMMENT                   // COMMENT
	LiteralBeg                // start of literals
	IDENT                     // main
	INT                       // 12345
	FLOAT                     // 123.45
	CHAR                      // 'a'
	STRING                    // "abc"
	LiteralEnd                // end of literals
	OperatorBeg               // start of operators and delimiters
	PLUS                      // +
	MINUS                     // -
	MULT                      // *
	DIV                       // /
	MODULO                    // %
	LPIPE                     // <<
	RPIPE                     // >>
	ACCUM                     // +=
	LOG_AND                   // &&
	LOG_OR                    // ||
	EQUAL                     // ==
	LESS                      // <
	GRTR                      // >
	ASSIGN                    // :=
	QASSIGN                   // ?=
	NOT                       // !
	NOT_EQUAL                 // !=
	LESS_EQUAL                // <=
	GRTR_EQUAL                // >=
	LPAREN                    // (
	LSQR                      // [
	LBRACE                    // {
	COMMA                     // ,
	PERIOD                    // .
	RPAREN                    // )
	RSQR                      // ]
	RBRACE                    // }
	SEMI                      // ;
	COLON                     // :
	DOLLAR                    // $
	DDOLLAR                   // $$
	OperatorEnd               // end of operators and delimiters
	KeywordBeg                // start of reserved/key words
	BREAK                     // break
	CONTINUE                  // continue
	ELSE                      // else
	FOR                       // for
	IN                        // in
	FUNC                      // func
	IF                        // if
	IMPORT                    // import
	PKG                       // pkg
	RETURN                    // return
	STRUCT                    // struct
	SWITCH                    // switch
	ISA                       // isa
	HASA                      // hasa
	TRUE                      // true
	FALSE                     // false
	WHILE                     // while
	NIL                       // nil
	ENUM                      // enum
	EXTERN                    // extern
	SYS                       // sys
	KeywordEnd                // end of reserved/key words
	TransformResultsBeg       // start of items produced by parser transforms
	FUNCAPPLY                 // function apply
	METHAPPLY                 // method apply
	STMTS                     //  statements
	TransformResultsEnd       // end of items produced by parser transforms
)

var tokens = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	COMMENT:    "COMMENT",
	IDENT:      "IDENT",
	INT:        "INT",
	FLOAT:      "FLOAT",
	CHAR:       "CHAR",
	STRING:     "STRING",
	PLUS:       "PLUS",
	MINUS:      "MINUS",
	MULT:       "MULT",
	DIV:        "DIV",
	MODULO:     "MODULO",
	LPIPE:      "LPIPE",
	RPIPE:      "RPIPE",
	ACCUM:      "ACCUM",
	LOG_AND:    "LOG_AND",
	LOG_OR:     "LOG_OR",
	EQUAL:      "EQUAL",
	LESS:       "LESS",
	GRTR:       "GRTR",
	ASSIGN:     "ASSIGN",
	QASSIGN:    "QASSIGN",
	NOT:        "NOT",
	NOT_EQUAL:  "NOT_EQUAL",
	LESS_EQUAL: "LESS_EQUAL",
	GRTR_EQUAL: "GRTR_EQUAL",
	LPAREN:     "LPAREN",
	LSQR:       "LSQR",
	LBRACE:     "LBRACE",
	COMMA:      "COMMA",
	PERIOD:     "PERIOD",
	RPAREN:     "RPAREN",
	RSQR:       "RSQR",
	RBRACE:     "RBRACE",
	SEMI:       "SEMI",
	COLON:      "COLON",
	DOLLAR:     "DOLLAR",
	DDOLLAR:    "DDOLLAR",
	BREAK:      "BREAK",
	CONTINUE:   "CONTINUE",
	ELSE:       "ELSE",
	FOR:        "FOR",
	IN:         "IN",
	FUNC:       "FUNC",
	IF:         "IF",
	IMPORT:     "IMPORT",
	PKG:        "PKG",
	RETURN:     "RETURN",
	STRUCT:     "STRUCT",
	SWITCH:     "SWITCH",
	ISA:        "ISA",
	HASA:       "HASA",
	TRUE:       "TRUE",
	FALSE:      "FALSE",
	WHILE:      "WHILE",
	NIL:        "NIL",
	ENUM:       "ENUM",
	EXTERN:     "EXTERN",
	SYS:        "SYS",
	FUNCAPPLY:  "FUNCAPPLY",
	METHAPPLY:  "METHAPPLY",
	STMTS:      "STMTS",
}

// String returns a string of a Token.
func (t Token) String() string {
	s := ""
	if 0 <= t && t < Token(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

var reserved = map[string]Token{
	"break":    BREAK,
	"continue": CONTINUE,
	"else":     ELSE,
	"for":      FOR,
	"in":       IN,
	"func":     FUNC,
	"if":       IF,
	"import":   IMPORT,
	"pkg":      PKG,
	"return":   RETURN,
	"struct":   STRUCT,
	"switch":   SWITCH,
	"isa":      ISA,
	"hasa":     HASA,
	"true":     TRUE,
	"false":    FALSE,
	"while":    WHILE,
	"nil":      NIL,
	"enum":     ENUM,
	"extern":   EXTERN,
	"sys":      SYS,
}

// CheckIdent checks if it's a reserved/keyword. Return either the reserved
// word token, or IDENT.
func CheckIdent(ident string) Token {

	if tok, ok := reserved[ident]; ok {
		return tok
	}

	return IDENT
}
