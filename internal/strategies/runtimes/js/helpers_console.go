package js

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type ConsoleLogMsg struct {
	logType string
	msg     string
}

func (jsr *JSRuntime) initConsole() error {
	return jsr.vm.Set("console", map[string]interface{}{
		"log": jsr.console_log,
	})
}

func (jsr *JSRuntime) console_log(args ...goja.Value) {
	if len(args) == 0 {
		return
	}

	parts := []string{}

	for _, p := range args {
		parts = append(parts, p.String())
	}

	msg := strings.Join(parts, " ")

	jsr.logs = append(jsr.logs, &ConsoleLogMsg{
		logType: "info",
		msg:     msg,
	})

	fmt.Printf("JSR: %+v\n", msg)
}
