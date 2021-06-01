package js

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (jsr *JSRuntime) initMath() error {
	jsr.vm.Set("math", mathLib())
	return jsr.vm.Set("Math", mathLib())
}

func mathLib() map[string]interface{} {
	base := map[string]interface{}{
		"abs":     abs,
		"acos":    acos,
		"acosh":   acosh,
		"asin":    asin,
		"asinh":   asinh,
		"atan":    atan,
		"atan2":   atan2,
		"atanh":   atanh,
		"cbrt":    cbrt,
		"ceil":    ceil,
		"clz32":   nosupport("clz32"),
		"cos":     cos,
		"cosh":    cosh,
		"exp":     exp,
		"expm1":   expm1,
		"floor":   floor,
		"fround":  fround,
		"hypot":   hypot,
		"imul":    nosupport("imul"),
		"log":     log,
		"log10":   log10,
		"log1p":   log1p,
		"log2":    log2,
		"max":     max,
		"min":     min,
		"pow":     pow,
		"random":  random,
		"round":   round,
		"sign":    sign,
		"sin":     sin,
		"sinh":    sinh,
		"sqrt":    sqrt,
		"tan":     tan,
		"tanh":    tanh,
		"trunc":   trunc,
		"E":       math.E,
		"LN10":    math.Ln10,
		"LOG10E":  math.Log10E,
		"LOG2E":   math.Log2E,
		"PI":      math.Pi,
		"SQRT2":   math.Sqrt2,
		"SQRT1_2": 0.7071067811865476,
	}

	for k, v := range base {
		base[strings.ToUpper(k)] = v
		base[strings.Title(k)] = v
	}

	return base
}

func nosupport(fn string) func() {
	return func() {
		panic(fmt.Sprintf("No support for this math call %s", fn))
	}
}

func log(v float64) float64 {
	return math.Log(v)
}

func abs(v float64) float64 {
	return math.Abs(v)
}

func acos(v float64) float64 {
	return math.Acos(v)
}

func acosh(v float64) float64 {
	return math.Acosh(v)
}

func asin(v float64) float64 {
	return math.Asin(v)
}

func asinh(v float64) float64 {
	return math.Asinh(v)
}

func atan(v float64) float64 {
	return math.Atan(v)
}

func atan2(a, b float64) float64 {
	return math.Atan2(a, b)
}

func atanh(v float64) float64 {
	return math.Atanh(v)
}

func cbrt(v float64) float64 {
	return math.Cbrt(v)
}

func ceil(v float64) float64 {
	return math.Ceil(v)
}

func cos(v float64) float64 {
	return math.Cos(v)
}

func cosh(v float64) float64 {
	return math.Cosh(v)
}

func exp(v float64) float64 {
	return math.Exp(v)
}

func expm1(v float64) float64 {
	return math.Expm1(v)
}

func floor(v float64) float64 {
	return math.Floor(v)
}

func fround(v float64) float64 {
	return float64(float32(v))
}

func hypot(a, b float64) float64 {
	return math.Hypot(a, b)
}

func log10(v float64) float64 {
	return math.Log10(v)
}

func log1p(v float64) float64 {
	return math.Log1p(v)
}

func log2(v float64) float64 {
	return math.Log2(v)
}

func max(a, b float64) float64 {
	return math.Max(a, b)
}

func min(a, b float64) float64 {
	return math.Min(a, b)
}

func pow(a, b float64) float64 {
	return math.Pow(a, b)
}

func random() float64 {
	return rand.Float64()
}

func round(v float64) float64 {
	return math.Round(v)
}

func sign(v float64) float64 {
	if v > 0 {
		return 1
	} else if v < 0 {
		return -1
	} else {
		if math.Signbit(v) {
			return 0
		}
		return -0
	}
}

func sin(v float64) float64 {
	return math.Sin(v)
}

func sinh(v float64) float64 {
	return math.Sinh(v)
}

func sqrt(v float64) float64 {
	return math.Sqrt(v)
}

func tan(v float64) float64 {
	return math.Tan(v)
}

func tanh(v float64) float64 {
	return math.Tanh(v)
}

func trunc(v float64) float64 {
	return math.Trunc(v)
}
