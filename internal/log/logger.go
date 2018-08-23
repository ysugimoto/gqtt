package log

import (
	"github.com/k0kubun/pp"
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

func Dump(v interface{}) {
	if isDebug {
		pp.Print(v)
	}
}
