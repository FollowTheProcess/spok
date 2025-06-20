package token_test

import (
	"fmt"
	"testing"

	"go.followtheprocess.codes/spok/token"
)

func TestToken_String(t *testing.T) {
	t.Parallel()
	type fields struct {
		Value string
		Type  token.Type
		Pos   int
		Line  int
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "error",
			want:   "Error message value",
			fields: fields{Value: "Error message value", Type: token.ERROR},
		},
		{
			name:   "eof",
			want:   "EOF",
			fields: fields{Value: "", Type: token.EOF},
		},
		{
			name:   "comment",
			want:   fmt.Sprintf("%q", "A comment"),
			fields: fields{Value: "A comment", Type: token.COMMENT},
		},
		{
			name:   "something long",
			want:   fmt.Sprintf("%q...", "A very very ver"),
			fields: fields{Value: "A very very very long comment", Type: token.COMMENT},
		},
		{
			name:   "hash",
			want:   fmt.Sprintf("%q", "#"),
			fields: fields{Value: "#", Type: token.HASH},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := token.Token{
				Value: tt.fields.Value,
				Type:  tt.fields.Type,
				Pos:   tt.fields.Pos,
				Line:  tt.fields.Line,
			}
			if got := tr.String(); got != tt.want {
				t.Errorf("Token.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_Is(t *testing.T) {
	t.Parallel()
	type fields struct {
		Value string
		Type  token.Type
		Pos   int
		Line  int
	}
	type args struct {
		typ token.Type
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "same type",
			fields: fields{Value: "A comment", Type: token.COMMENT},
			args:   args{typ: token.COMMENT},
			want:   true,
		},
		{
			name:   "different type",
			fields: fields{Value: "A comment", Type: token.COMMENT},
			args:   args{typ: token.TASK},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := token.Token{
				Value: tt.fields.Value,
				Type:  tt.fields.Type,
				Pos:   tt.fields.Pos,
				Line:  tt.fields.Line,
			}
			if got := tr.Is(tt.args.typ); got != tt.want {
				t.Errorf("Token.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
		i    token.Type
	}{
		{
			name: "error",
			want: "ERROR",
			i:    token.ERROR,
		},
		{
			name: "left paren",
			want: "(",
			i:    token.LPAREN,
		},
		{
			name: "task",
			want: "task",
			i:    token.TASK,
		},
		{
			name: "declare",
			want: ":=",
			i:    token.DECLARE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.String(); got != tt.want {
				t.Errorf("Type.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
