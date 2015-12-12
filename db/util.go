package crdb

import (
    "log"
    "runtime"
    "strconv"
)

func __log(t, m string, v ...interface{}) {
    pc, _, ln, _ := runtime.Caller(2)
    fn := runtime.FuncForPC(pc).Name()
    log.Printf(fn + ":" + strconv.Itoa(ln) + " -- " + t + " -- " + m + "\n", v...)
}

func LogInfo(m string, v ...interface{}) { __log("INFO", m, v...) }
func LogWarn(m string, v ...interface{}) { __log("INFO", m, v...) }
func LogError(m string, v ...interface{}) { __log("INFO", m, v...) }

