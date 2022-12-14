package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
)

func TestSimpleRun(t *testing.T) {
	jsr := &JSRuntime{}
	err := jsr.Init(
		[]byte(`
			return SELL;
		`),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	v, err := jsr.Run()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v, strategy.Action_SELL)
}

func TestInnerPanic(t *testing.T) {
	jsr := &JSRuntime{
		enableTestSuite: true,
	}
	err := jsr.Init(
		[]byte(`
			timeout();
		`),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = jsr.Run()

	assert.ErrorIs(t, err, ErrTimeout)
}

func TestRunTradesPanic(t *testing.T) {
	jsr := &JSRuntime{
		enableTestSuite: true,
	}
	err := jsr.Init(
		[]byte(`
			let tr = GetTrades('binance.com', 'ADAAUD', '5m');
			var last = 0; var sum = 0.0;

			tr.forEach(trade => {
				if (last == 0) {
					last = trade.Amount
					return
				}

				sum += math.log10(trade.Amount/last);
				last = trade.Amount;
			});

			console.log(sum);

			if (sum >=0.001) {
				return BUY;
			}

			return SELL;
		`),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	v, err := jsr.Run()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v, strategy.Action_BUY)
	assert.NotEmpty(t, jsr.logs)
}

func TestCandles(t *testing.T) {
	jsr := &JSRuntime{
		enableTestSuite: true,
	}
	err := jsr.Init(
		[]byte(`
			let data = GetCandles('binance.com', 'ADAAUD', '1h', 10);

			let last = data.length - 1

			let momd = (data[last].Close/data[last-5].Close) * 100
			let momd40 = (data[last].Close/data[last-15].Close) * 100

			console.log(momd, momd40)

			if (momd <= momd40) {
				return SELL;
			}

			return BUY;
		`),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	v, err := jsr.Run()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v, strategy.Action_BUY)
	assert.NotEmpty(t, jsr.logs)
}
