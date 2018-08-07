package log

import (
	. "log"
	"os"
)

var isDebug bool

func init() {
	isDebug = os.Getenv("DEBUG") != ""
}

func Debug(v ...interface{}) {
	if isDebug {
		Println(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if isDebug {
		Printf(format, v...)
	}
}
