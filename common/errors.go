package common

import "errors"

var (
	ErrScriptNotExist = errors.New("given script file doesn't exist")
	ErrScriptExecError = errors.New("failed to runner the given script file")
)
