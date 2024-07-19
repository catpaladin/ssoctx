package main

import (
	"reflect"
	"testing"

	"github.com/rs/zerolog"
)

func TestConfigureLogger(t *testing.T) {
	type args struct {
		debug      bool
		jsonFormat bool
	}
	tests := []struct {
		name string
		args args
		want zerolog.Level
	}{
		{
			name: "Create Logger all true",
			args: args{
				debug:      true,
				jsonFormat: true,
			},
			want: zerolog.TraceLevel,
		},
		{
			name: "Create Logger all false",
			args: args{
				debug:      false,
				jsonFormat: false,
			},
			want: zerolog.TraceLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := configureLogger(tt.args.debug, tt.args.jsonFormat); !reflect.DeepEqual(got.GetLevel(), tt.want) {
				t.Errorf("ConfigureLogger() = %v, want %v", got.GetLevel(), tt.want)
			}
		})
	}
}
