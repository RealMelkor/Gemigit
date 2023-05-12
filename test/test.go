package test

import (
	"testing"
	"runtime"
	"strings"
	"strconv"
)

func fileAndLine() string {
        _, file, no, ok := runtime.Caller(2)
        if !ok {
                return ""
        }
        path := strings.Split(file, "/")
        return path[len(path) - 1] + ":" + strconv.Itoa(no) + ":"
}

func IsNil(t *testing.T, err error) {
        if err != nil {
                t.Fatal(fileAndLine(), err)
        }
}

func IsNotNil(t *testing.T, err error, message string) {
        if err == nil {
                t.Fatal(fileAndLine(), message)
        }
}

func IsEqual(t *testing.T, x interface{}, y interface{}) {
        if x != y {
                t.Fatal(fileAndLine(), x, " != ", y)
        }
}

func IsNotEqual(t *testing.T, x interface{}, y interface{}) {
        if x == y {
                t.Fatal(fileAndLine(), x, " != ", y)
        }
}

func FuncName(t *testing.T) string {
        fpcs := make([]uintptr, 1)

        n := runtime.Callers(2, fpcs)
        if n == 0 {
                t.Fatal("function name: no caller")
        }

        caller := runtime.FuncForPC(fpcs[0] - 1)
        if caller == nil {
                t.Fatal("function name: caller is nil")
        }

        name := caller.Name()
        return name[strings.LastIndex(name, ".") + 1:]
}
