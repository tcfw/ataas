package js

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
)

var (
	ErrCodeTooLarge = errors.New("code block larger than allowed execution size")
	ErrTimeout      = errors.New("timed out")
)

type PanicErr struct {
	err error
}

func (e *PanicErr) Error() string {
	return fmt.Sprintf("panic err: %s", e.err)
}

type JSRuntime struct {
	code            string
	params          map[string]string
	vm              *goja.Runtime
	enableTestSuite bool

	logs []*ConsoleLogMsg
}

func (jsr *JSRuntime) Init(code []byte, params map[string]string) error {
	jsr.logs = make([]*ConsoleLogMsg, 0)

	if !strings.HasPrefix(string(code), `(function(){`) &&
		!strings.HasSuffix(string(code), `})();`) {
		code = []byte(fmt.Sprintf(`(function(){%s})();`, code))
	}

	code, err := convertJS(code)
	if err != nil {
		return err
	}

	jsr.vm = goja.New()
	jsr.code = string(code)
	jsr.params = params

	if len(code) > 1<<10 {
		return ErrCodeTooLarge
	}

	err = jsr.initParams()
	if err != nil {
		return err
	}

	return nil
}

func (jsr *JSRuntime) initParams() error {
	for k, v := range jsr.params {
		err := jsr.vm.Set(k, v)
		if err != nil {
			return err
		}
	}

	jsr.vm.Set("BUY", strategy.Action_BUY)
	jsr.vm.Set("SELL", strategy.Action_SELL)
	jsr.vm.Set("STAY", strategy.Action_STAY)

	if jsr.enableTestSuite {
		err := jsr.vm.Set("timeout", func() {
			panic(ErrTimeout)
		})
		if err != nil {
			return err
		}
		err = jsr.vm.Set("GetTrades", GetTestTrades)
		if err != nil {
			return err
		}
	} else {
		err := jsr.vm.Set("GetTrades", GetTrades)
		if err != nil {
			return err
		}
	}

	if err := jsr.initMath(); err != nil {
		return err
	}

	if err := jsr.initConsole(); err != nil {
		return err
	}

	return nil
}

func (jsr *JSRuntime) Run() (strat strategy.Action, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			if caught == ErrTimeout {
				err = ErrTimeout
				strat = strategy.Action_STAY
				return
			}

			if e, ok := caught.(error); ok {
				strat = strategy.Action_STAY
				err = &PanicErr{e}
				return
			}

			panic(caught) // Something else happened, repanic!
		}
	}()

	go func() {
		time.Sleep(30 * time.Second) // Stop after 30 seconds
		jsr.vm.Interrupt(ErrTimeout)
	}()

	v, err := jsr.vm.RunString(jsr.code)
	if err != nil {
		return strategy.Action_STAY, err
	}

	retV64 := v.ToInteger()
	if err != nil {
		return strategy.Action_STAY, err
	}

	return strategy.Action(retV64), nil
}
