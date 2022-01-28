// Code generated by "stringer -type=Type -linecomment -output=token_string.go"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ERROR-0]
	_ = x[EOF-1]
	_ = x[COMMENT-2]
	_ = x[HASH-3]
	_ = x[LPAREN-4]
	_ = x[RPAREN-5]
	_ = x[LBRACE-6]
	_ = x[RBRACE-7]
	_ = x[LBRACKET-8]
	_ = x[RBRACKET-9]
	_ = x[QUOTE-10]
	_ = x[COMMA-11]
	_ = x[TASK-12]
	_ = x[STRING-13]
	_ = x[COMMAND-14]
	_ = x[INTEGER-15]
	_ = x[OUTPUT-16]
	_ = x[IDENT-17]
	_ = x[DECLARE-18]
}

const _Type_name = "ERROREOFCOMMENT#(){}[]\",taskSTRINGCOMMANDINTEGER->IDENT:="

var _Type_index = [...]uint8{0, 5, 8, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 28, 34, 41, 48, 50, 55, 57}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
