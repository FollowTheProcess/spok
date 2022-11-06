package shell_test

import (
	"testing"

	"github.com/FollowTheProcess/spok/iostream"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/google/go-cmp/cmp"
)

func TestResultOk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		result shell.Result
		want   bool
	}{
		{
			name:   "yes",
			result: shell.Result{Status: 0},
			want:   true,
		},
		{
			name:   "no",
			result: shell.Result{Status: 1},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Ok(); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestResultsOk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		result shell.Results
		want   bool
	}{
		{
			name: "all ok",
			result: shell.Results{
				{Status: 0},
				{Status: 0},
				{Status: 0},
				{Status: 0},
			},
			want: true,
		},
		{
			name: "first bad",
			result: shell.Results{
				{Status: 1},
				{Status: 0},
				{Status: 0},
				{Status: 0},
			},
			want: false,
		},
		{
			name: "last bad",
			result: shell.Results{
				{Status: 0},
				{Status: 0},
				{Status: 0},
				{Status: 3},
			},
			want: false,
		},
		{
			name: "middle bad",
			result: shell.Results{
				{Status: 0},
				{Status: 2},
				{Status: 0},
				{Status: 0},
			},
			want: false,
		},
		{
			name: "several bad",
			result: shell.Results{
				{Status: 0},
				{Status: 2},
				{Status: 1},
				{Status: 3},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Ok(); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cmd     string
		env     []string
		want    shell.Result
		wantErr bool
	}{
		{
			name: "echo",
			cmd:  "echo hello",
			want: shell.Result{
				Stdout: "hello\n",
				Stderr: "",
				Status: 0,
				Cmd:    "echo hello",
			},
			wantErr: false,
		},
		{
			name: "echo no newline",
			cmd:  "echo -n hello",
			want: shell.Result{
				Stdout: "hello",
				Stderr: "",
				Status: 0,
				Cmd:    "echo -n hello",
			},
			wantErr: false,
		},
		{
			name: "echo stderr",
			cmd:  "echo This message goes to stderr >&2",
			want: shell.Result{
				Stdout: "",
				Stderr: "This message goes to stderr\n",
				Status: 0,
				Cmd:    "echo This message goes to stderr >&2",
			},
			wantErr: false,
		},
		{
			name: "exit 0",
			cmd:  "exit 0",
			want: shell.Result{
				Stdout: "",
				Stderr: "",
				Status: 0,
				Cmd:    "exit 0",
			},
			wantErr: false,
		},
		{
			name: "exit 1",
			cmd:  "exit 1",
			want: shell.Result{
				Stdout: "",
				Stderr: "",
				Status: 1,
				Cmd:    "exit 1",
			},
			wantErr: false,
		},
		{
			name: "environment",
			cmd:  "echo $VARIABLE",
			env:  []string{"VARIABLE=hello"},
			want: shell.Result{
				Stdout: "hello\n",
				Stderr: "",
				Status: 0,
				Cmd:    "echo $VARIABLE",
			},
			wantErr: false,
		},
		{
			name: "bad syntax",
			cmd:  "(*^$$",
			want: shell.Result{
				Stdout: "",
				Stderr: "",
				Status: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := shell.NewIntegratedRunner()
			got, err := runner.Run(tt.cmd, iostream.Null(), tt.name, tt.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Run() err = %v, wantErr = %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
