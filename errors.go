package redis_orm

import "fmt"

type ErrorWithCode struct {
	msg  string
	code int
}

func (e ErrorWithCode) Error() string {
	return e.msg
}

func (e ErrorWithCode) Code() int {
	return e.code
}

func (e ErrorWithCode) Append(format string, a ...interface{}) *ErrorWithCode {
	if len(a) > 0 {
		return Error(e.Code(), e.msg+" "+fmt.Sprintf(format, a))
	} else {
		return Error(e.Code(), e.msg+" "+format)
	}
}

func (e *ErrorWithCode) Equal(err error) bool {
	if errWithCode, ok := err.(ErrorWithCode); ok {
		return errWithCode.Code() == e.Code()
	} else {
		return false
	}
}
func Code(err error) int {
	if err == nil {
		return ErrorCode_Success
	}
	if errWithCode, ok := err.(ErrorWithCode); ok {
		return errWithCode.Code()
	}
	return ErrorCode_Unexpected
}
func Error(code int, format string, a ...interface{}) *ErrorWithCode {
	err := &ErrorWithCode{
		code: code,
		msg:  fmt.Sprintf(format, a...),
	}
	if len(a) == 0 {
		err.msg = format
	}
	return err
}
