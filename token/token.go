package token

// TokenType is what particular type of token found.
type TokenType string

// Token contains the type and literal value of a parsed token.
type Token struct {
	Type    TokenType
	Literal string
}

// token values
const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"   // 1343456

	// Operators
	ASSIGN     = ":="
	NIL_ASSIGN = "?="
	PLUS       = "+"
	MINUS      = "-"
	BANG       = "!"
	ASTERISK   = "*"
	SLASH      = "/"
	AND        = "&&"
	OR         = "||"
	XOR        = "^^"

	EQ     = "=="
	NOT_EQ = "!="
	LT     = "<"
	GT     = ">"
	LT_EQ  = "<="
	GT_EQ  = ">="

	// pipe
	PIPE = ">>"

	// magic marker
	TILDE = "~"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	LSQR   = "["
	RSQR   = "]"

	// Keywords
	FUNC   = "FUNC"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	IF     = "IF"
	ELSE   = "ELSE"
	RETURN = "RETURN"
	YIELD  = "YIELD"
	WHILE  = "WHILE"
	FOR    = "FOR"
	IN     = "IN"
	NIL    = "NIL"
	STRUCT = "STRUCT"
	SWITCH = "SWITCH"
	CASE   = "CASE"
	ISA    = "ISA"
	HASA   = "HASA"
	ENUM   = "ENUM"
	IMPORT = "IMPORT"
	PKG    = "PKG"
	SYS    = "SYS"
)

var keywords = map[string]TokenType{
	"func":   FUNC,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"yield":  YIELD,
	"while":  WHILE,
	"for":    FOR,
	"in":     IN,
	"nil":    NIL,
	"struct": STRUCT,
	"switch": SWITCH,
	"case":   CASE,
	"isa":    ISA,
	"hasa":   HASA,
	"enum":   ENUM,
	"import": IMPORT,
	"pkg":    PKG,
	"sys":    SYS,
}

// LookupIdent checks if it's a reserved/keyword. Return either the reserved
// word token, or IDENT.
func LookupIdent(ident string) TokenType {

	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
