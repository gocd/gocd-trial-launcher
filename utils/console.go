package utils

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
)

func Debug(f string, t ...interface{}) {
	if EnableDebug {
		Out(fmt.Sprintf(`[DEBUG] %s`, f), t...)
	}
}

func Out(f string, t ...interface{}) {
	if len(t) > 0 {
		_, _ = fmt.Fprintf(colorable.NewColorableStdout(), f+"\n", t...)
	} else {
		_, _ = fmt.Fprint(colorable.NewColorableStdout(), f+"\n")
	}
}

func Err(f string, t ...interface{}) {
	if len(t) > 0 {
		_, _ = fmt.Fprintf(colorable.NewColorableStderr(), f+"\n", t...)
	} else {
		_, _ = fmt.Fprint(colorable.NewColorableStderr(), f+"\n")
	}
}

// Die Exits with exitCode after printing message.
// Automatically selects STDOUT vs STDERR depending
// on value of exitCode
func Die(exitCode int, f string, t ...interface{}) {
	if exitCode != 0 {
		Err(f, t...)
	} else {
		Out(f, t...)
	}

	os.Exit(exitCode)
}
