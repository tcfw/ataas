package runtimes

import (
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
)

type Runtime interface {
	Init(code []byte, params map[string]string) error
	Run() (strategy.Action, error)
}
