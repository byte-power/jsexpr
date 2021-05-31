package builtin

import (
	"math"
	"math/bits"
	"math/rand"
	"time"
)

var objects = map[string]interface{}{
	"Date": dateObject{
		Now: time.Now().Unix,
	},
	"Math": mathObject{
		E:       E,
		LN2:     LN2,
		LN10:    LN10,
		LOG2E:   LOG2E,
		LOG10E:  LOG10E,
		PI:      PI,
		SQRT1_2: SQRT1_2,
		SQRT2:   SQRT2,

		Abs:   math.Abs,
		Acos:  math.Acos,
		Acosh: math.Acosh,
		Asin:  math.Asin,
		Asinh: math.Asinh,
		Atan:  math.Atan,
		Atan2: math.Atan2,
		Atanh: math.Atanh,

		Cbrt:  math.Cbrt,
		Ceil:  math.Ceil,
		Clz32: bits.LeadingZeros32,
		Cos:   math.Cos,
		Cosh:  math.Cosh,
		Exp:   math.Exp,
		ExpM1: math.Expm1,
		Floor: math.Floor,
		//Fround: ,

		Hypot: jsHypotenuse,

		// Imul: bits.mu,

		Log:   math.Log,
		Log1p: math.Log1p,
		Log2:  math.Log2,
		Log10: math.Log10,

		Max: jsMax,
		Min: jsMin,

		Pow: math.Pow,

		Random: rand.Float64,
		Round:  math.Round,

		Sign: jsMathSign,
		Sin:  math.Sin,
		Sinh: math.Sinh,
		Sqrt: math.Sqrt,

		Tan:   math.Tan,
		Tanh:  math.Tanh,
		Trunc: math.Trunc,
	},
}

func Objs() map[string]interface{} {
	return objects
}

type dateObject struct {
	Now func() int64 `jsexpr:"now"`
}

var (
	E       float64 = math.E
	LN2     float64 = math.Ln2
	LN10    float64 = math.Ln10
	LOG2E   float64 = math.Log2E
	LOG10E  float64 = math.Log10E
	PI      float64 = math.Pi
	SQRT1_2 float64 = math.Sqrt(0.5)
	SQRT2   float64 = math.Sqrt2
)

type mathObject struct {
	E       float64 `jsexpr:"E"`
	LN2     float64 `jsexpr:"LN2"`
	LN10    float64 `jsexpr:"LN10"`
	LOG2E   float64 `jsexpr:"LOG2E"`
	LOG10E  float64 `jsexpr:"LOG10E"`
	PI      float64 `jsexpr:"PI"`
	SQRT1_2 float64 `jsexpr:"SQRT1_2"`
	SQRT2   float64 `jsexpr:"SQRT2"`

	Abs   func(x float64) float64    `jsexpr:"abs"`
	Acos  func(x float64) float64    `jsexpr:"acos"`
	Acosh func(x float64) float64    `jsexpr:"acosh"`
	Asin  func(x float64) float64    `jsexpr:"asin"`
	Asinh func(x float64) float64    `jsexpr:"asinh"`
	Atan  func(x float64) float64    `jsexpr:"atan"`
	Atan2 func(y, x float64) float64 `jsexpr:"atan2"`
	Atanh func(x float64) float64    `jsexpr:"atanh"`

	Cbrt  func(x float64) float64 `jsexpr:"cbrt"`
	Ceil  func(x float64) float64 `jsexpr:"ceil"`
	Clz32 func(x uint32) int      `jsexpr:"clz32"`
	Cos   func(x float64) float64 `jsexpr:"cos"`
	Cosh  func(x float64) float64 `jsexpr:"cosh"`

	Exp   func(x float64) float64 `jsexpr:"exp"`
	ExpM1 func(x float64) float64 `jsexpr:"expm1"`

	Floor  func(x float64) float64 `jsexpr:"floor"`
	Fround func(x float64) float64 `jsexpr:"fround"`

	Hypot func(nums ...float64) float64 `jsexpr:"hypot"`

	Imul func(x, y float64) float64 `jsexpr:"imul"`

	Log   func(x float64) float64 `jsexpr:"log"`
	Log1p func(x float64) float64 `jsexpr:"log1p"`
	Log2  func(x float64) float64 `jsexpr:"log2"`
	Log10 func(x float64) float64 `jsexpr:"log10"`

	Max func(nums ...float64) float64 `jsexpr:"max"`
	Min func(nums ...float64) float64 `jsexpr:"min"`

	Pow func(x, y float64) float64 `jsexpr:"pow"`

	Random func() float64          `jsexpr:"random"`
	Round  func(x float64) float64 `jsexpr:"round"`

	Sign func(x float64) int     `jsexpr:"sign"`
	Sin  func(x float64) float64 `jsexpr:"sin"`
	Sinh func(x float64) float64 `jsexpr:"sin"`
	Sqrt func(x float64) float64 `jsexpr:"sqrt"`

	Tan   func(x float64) float64 `jsexpr:"tan"`
	Tanh  func(x float64) float64 `jsexpr:"tanh"`
	Trunc func(x float64) float64 `jsexpr:"trunc"`
}

func jsMathSign(x float64) int {
	if x == 0 {
		return 0
	}
	if x > 0 {
		return 1
	}
	return -1
}

//弦
func jsHypotenuse(nums ...float64) float64 {
	sum := float64(0)
	for _, num := range nums {
		sum += math.Pow(num, 2)
	}
	return math.Sqrt(sum)
}

func jsMax(nums ...float64) float64 {
	pivot := math.Inf(-1)
	for _, num := range nums {
		if num > pivot {
			pivot = num
		}
	}
	return pivot
}

func jsMin(nums ...float64) float64 {
	pivot := math.Inf(+1)
	for _, num := range nums {
		if num < pivot {
			pivot = num
		}
	}
	return pivot
}
