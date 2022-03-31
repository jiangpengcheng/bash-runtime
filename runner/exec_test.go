package runner

import (
	"bash-runtime/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExec(t *testing.T) {
	tests := []struct {
		name         string
		script       string
		param        string
		expectStdout string
		expectStderr string
		expectError  error
	}{
		{
			name:         "it should runner the script correctly",
			script:       "../scripts/exec.sh",
			param:        "hello world",
			expectStdout: "hello world!",
			expectStderr: "",
			expectError:  nil,
		},
		{
			name:         "it should return error when runner a non-exist script",
			script:       "../scripts/non-exist.sh",
			param:        "hello world",
			expectStdout: "<nil>",
			expectStderr: "<nil>",
			expectError:  common.ErrScriptNotExist,
		},
		{
			name:         "it should return error when runner a non-executable script",
			script:       "../scripts/non-executable.sh",
			param:        "hello world",
			expectStdout: "<nil>",
			expectStderr: "<nil>",
			expectError:  common.ErrScriptNotExist,
		},
		{
			name:         "it should get the output of stderr",
			script:       "../scripts/stderr.sh",
			param:        "hello world",
			expectStdout: "hello world!",
			expectStderr: "../scripts/stderr.sh: line 3: data: command not found\n",
			expectError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := execScript(tt.script, tt.param)
			assert.Equal(t, tt.expectStdout, stdout.String())
			assert.Equal(t, tt.expectStderr, stderr.String())
			assert.Equal(t, tt.expectError, err)
		})
	}
}
