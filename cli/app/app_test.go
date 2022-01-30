package app

import (
	"bytes"
	"io"
	"testing"
)

func TestApp_Run(t *testing.T) {
	type fields struct {
		Out     io.Writer
		Options *Options
	}
	type args struct {
		tasks []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "just tasks",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{},
			},
			args:    args{tasks: []string{"task1", "task2", "task3"}},
			wantErr: false,
		},
		{
			name: "no tasks",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{},
			},
			args:    args{tasks: []string{}},
			wantErr: false,
		},
		{
			name: "show",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{Show: "task1"},
			},
			args:    args{tasks: []string{}},
			wantErr: false,
		},
		{
			name: "fmt",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{Fmt: true},
			},
			args:    args{tasks: []string{}},
			wantErr: false,
		},
		{
			name: "variables",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{Variables: true},
			},
			args:    args{tasks: []string{}},
			wantErr: false,
		},
		{
			name: "spokfile",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{Spokfile: "dinglefile"},
			},
			args:    args{tasks: []string{}},
			wantErr: false,
		},
		{
			name: "clean",
			fields: fields{
				Out:     &bytes.Buffer{},
				Options: &Options{Clean: true},
			},
			args:    args{tasks: []string{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				Out:     tt.fields.Out,
				Options: tt.fields.Options,
			}
			if err := a.Run(tt.args.tasks); (err != nil) != tt.wantErr {
				t.Errorf("App.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
