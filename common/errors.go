package common

import "errors"

var (
	ErrScriptNotExist = errors.New("given script file doesn't exist")
	ErrScriptExecError = errors.New("failed to run the given script file")
)
