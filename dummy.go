package main

import (
	"fmt"

	"pm.tcfw.com.au/source/trader/common"
)

type DummyMarketController struct{}

var _ common.OrderController = &DummyMarketController{}

func (dmc *DummyMarketController) Buy(units float64, price float64) (float64, error) {
	fmt.Println("DUMMY BUY:", units, price)
	return units, nil
}

func (dmc *DummyMarketController) Sell(units float64) error {
	fmt.Println("DUMMY SELL:", units)
	return nil
}
