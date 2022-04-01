package common

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	noEnv := "nil"
	type args struct {
		key      string
		fallback string
	}
	tests := []struct {
		name string
		args args
		env  string
		want string
	}{
		{
			name: "it should get env",
			args: args{
				key: "TEST",
				fallback: "default value",
			},
			env: "test value",
			want: "test value",
		},
		{
			name: "it should get env when it's empty",
			args: args{
				key: "TEST",
				fallback: "default value",
			},
			env: "",
			want: "",
		},
		{
			name: "it should get fallback when env is not set",
			args: args{
				key: "TEST",
				fallback: "default value",
			},
			env: noEnv,
			want: "default value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != noEnv {
				os.Setenv(tt.args.key, tt.env)
			}
			if got := GetEnv(tt.args.key, tt.args.fallback); got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
			if tt.env != noEnv {
				os.Unsetenv(tt.args.key)
			}
		})
	}
}
