package binance

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFees(t *testing.T) {
	c := NewClient("GJPXJLLoVGxXsxoeh9kbb8aBd9LSWkjIfNTqeMNHarkAj6UmRKNDL2MS9Ul1HQRN", "5lkciHfxXu08cRNb74Xxl1hb7D5QugiOBeTQI9D5TPZsnZhkd8Tl9jxFaGxShEYD")

	make, take, err := c.fees("ADAAUD")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotZero(t, make)
	assert.NotZero(t, take)
}

func TestTrunc(t *testing.T) {
	f, _ := strconv.ParseFloat("0.0123456789", 32)

	assert.Equal(t, truncatePrecision(f, 3), 0.012)
	assert.Equal(t, truncatePrecision(f, 4), 0.0123)
	assert.Equal(t, truncatePrecision(f, 8), 0.01234567)
}
