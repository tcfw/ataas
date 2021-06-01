package strategies

import (
	"encoding/json"
	"fmt"

	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/internal/strategies/runtimes/js"
)

func (w *Worker) handleJSRuntime(job *strategy.Strategy) error {
	jsparams := map[string]string{}
	if p, ok := job.Params["params"]; ok {
		err := json.Unmarshal([]byte(p), &jsparams)
		if err != nil {
			return err
		}
	}

	code, ok := job.Params["code"]
	if !ok {
		return fmt.Errorf("no code in strategy")
	}

	jsr := &js.JSRuntime{}
	err := jsr.Init([]byte(code), jsparams)
	if err != nil {
		return err
	}

	action, err := jsr.Run()
	if err != nil {
		return err
	}

	err = w.storeSuggestedAction(action, job)
	if err != nil {
		return err
	}

	return w.broadcastSuggestedAction(action, job)
}
