// Code generated by "stringer -type=Type -linecomment -output=token_string.go"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EOF-0]
	_ = x[ERROR-1]
	_ = x[COMMENT-2]
	_ = x[HASH-3]
	_ = x[LPAREN-4]
	_ = x[RPAREN-5]
	_ = x[LBRACE-6]
	_ = x[RBRACE-7]
	_ = x[QUOTE-8]
	_ = x[COMMA-9]
	_ = x[TASK-10]
	_ = x[STRING-11]
	_ = x[COMMAND-12]
	_ = x[OUTPUT-13]
	_ = x[IDENT-14]
	_ = x[DECLARE-15]
	_ = x[LINTERP-16]
	_ = x[RINTERP-17]
}

const _Type_name = "EOFERRORCOMMENT#(){}\",taskSTRINGCOMMAND->IDENT:={{}}"

var _Type_index = [...]uint8{0, 3, 8, 15, 16, 17, 18, 19, 20, 21, 22, 26, 32, 39, 41, 46, 48, 50, 52}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
