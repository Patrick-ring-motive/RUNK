package RUNK

/*Ring Universal Number Kounter*/
/*
The RUNK package provides generic alternatives to the std math package to allow usage
with any of go's number types. There are some minor additions added for convenience but the original
behavior is preserved.
*/
import (
	"github.com/Patrick-ring-motive/utils"
	"math"
	"strconv"
)

/*
Some notes on some unusual patterns that I employ.

`func(T){}` is a function that takes a single typed parameter and does nothing.
This is my way of passing a type reference around without having to instantiate it.
Mostly it is abstracted away but you will see it show in parameters sometimes.
Utils has a convenient function to do this called TypeRef which you can use to generate these references.
You can generate them from an abstract type like so: `utils.TypeRef[int]()` or from a concrete type like so: `utils.TypeRef(0)`.

`*[1]T` is a pointer to an array of length 1 of type T. This ensures that values are passed by reference and not
copied and facilitates passing values back from a defer/recover block.

That brings me to the unconventional error handling pattern.

```
func Example(input inputType)outputType{
  var z outputType
  carrier := *[1]outputType{z}
  example(carrier,input)
  return carrier[0]
}
func example(carrier *[1]outputType,input inputType){
	defer func() {
		if r := recover(); r != nil {
			carrier[0] = fallbackValue
		}
	}()
	carrier[0] = attemptSomething(input)
	if(carrier[0] == nil){
		carrier[0] = fallbackValue
	}
}
```

This is a pattern than handles errors by "returning" a fallback value
on panic or nil. This pattern is difficult to abstract out because go generics dont handle various function types well and the defer needs to happen one function call deeper than where we intend to recover a panic.
*/

/*Constant values for minimum and maximum values. Most are directly ripped from the original math*/
const (
	MaxInt                 int     = 1<<(strconv.IntSize-1) - 1  // MaxInt32 or MaxInt64 depending on intSize.
	MinInt                 int     = -1 << (strconv.IntSize - 1) // MinInt32 or MinInt64 depending on intSize.
	MaxInt8                int8    = 1<<7 - 1                    // 127
	MinInt8                int8    = -1 << 7                     // -128
	MaxInt16               int16   = 1<<15 - 1                   // 32767
	MinInt16               int16   = -1 << 15                    // -32768
	MaxInt32               int32   = 1<<31 - 1                   // 2147483647
	MinInt32               int32   = -1 << 31                    // -2147483648
	MaxInt64               int64   = 1<<63 - 1                   // 9223372036854775807
	MinInt64               int64   = -1 << 63                    // -9223372036854775808
	MaxUint                uint    = 1<<strconv.IntSize - 1      // MaxUint32 or MaxUint64 depending on intSize.
	MinUint                uint    = 0
	MaxUint8               uint8   = 1<<8 - 1 // 255
	MinUint8               uint8   = 0
	MaxUint16              uint16  = 1<<16 - 1 // 65535
	MinUint16              uint16  = 0
	MaxUint32              uint32  = 1<<32 - 1 // 4294967295
	MinUint32              uint32  = 0
	MaxUint64              uint64  = 1<<64 - 1 // 18446744073709551615
	MinUint64              uint64  = 0
	MaxUintptr             uintptr = uintptr(MaxUint)
	MinUintptr             uintptr = 0
	MaxByte                byte    = byte(MaxUint8)
	MinByte                byte    = 0
	MaxFloat32             float32 = 0x1p127 * (1 + (1 - 0x1p-23)) // 3.40282346638528859811704183484516925440e+38
	MinFloat32             float32 = -MaxFloat32
	SmallestNonzeroFloat32 float32 = 0x1p-126 * 0x1p-23             // 1.401298464324817070923729583289916131280e-45
	MaxFloat64             float64 = 0x1p1023 * (1 + (1 - 0x1p-52)) // 1.79769313486231570814527423731704356798070e+308
	MinFloat64             float64 = -MaxFloat64
	SmallestNonzeroFloat64 float64 = 0x1p-1022 * 0x1p-52 // 4.9406564584124654417656879286822137236505980e-324
)

type runk string

var Runk = runk("RUNK")

/*
byte is an alias for uint8 but it seems wrong to not include it explicitly.
The compiler won't let me do have both but just so you know it isn't forgotten I have included it with this interface.
Also rune isn't included because I think it should be treated more like a character than a number. That may change. I also considered having boolean and string representations of numbers but for now these will do.
*/
type ibyte interface {
	~byte
}

/* All the numbers! */
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ibyte
}

/*
AsNumber is just used as a compiler hint to tell the compiler that the value can be any Number type.
This is used to make the compiler happy when using narrowing conversions.
*/
func AsNumber[N Number](n N) N {
	return n
}

/*
ConvertNum is the most flexible conversion function as it accepts an any type.
It is needed to make number conversion more concise. Even though it is intended to use with numbers, it will make a best effort to convert non number types. Typical usade looks like
`ConvertNum[int](11.2)` which will return 11.
*/
func ConvertNum[To Number](f any, t ...func(To)) To {
	switch v := f.(type) {
	case int:
		return ConvertNumber[int, To](v)
	case int8:
		return ConvertNumber[int8, To](v)
	case int16:
		return ConvertNumber[int16, To](v)
	case int32:
		return ConvertNumber[int32, To](v)
	case int64:
		return ConvertNumber[int64, To](v)
	case uint:
		ConvertNumber[uint, To](v)
	case uint8:
		return ConvertNumber[uint8, To](v)
	case uint16:
		return ConvertNumber[uint16, To](v)
	case uint32:
		return ConvertNumber[uint32, To](v)
	case uint64:
		return ConvertNumber[uint64, To](v)
	case uintptr:
		return ConvertNumber[uintptr, To](v)
	case float32:
		return ConvertNumber[float32, To](v)
	case float64:
		return ConvertNumber[float64, To](v)
	case bool:
		if v {
			return To(1)
		} else {
			return To(0)
		}
	default:
		return utils.ConvertType[any, To](f)
	}
	return utils.ConvertType[any, To](f)
}

/*
This is the helper function that makes all this possible without having to write a different implementation
for each indiviual type. ConvertNumber handles converting between number types while maintaining the generic
Number designation on the return. You'll see me using the `func(T)` syntax which is a pattern that I use to
pass around a type without having to instantiate it. This works to give the compiler a hint that it can use to
maitain type safety. This should work for most scenarios.
The main edge cases to worry about are NaN and Inf which can get coerced into a number that isn't very meaningful. NaN converted to an int will return 0 so that at least it maintains the same truthiness and +/- Inf converted to an int will return MaxInt/MinInt. In narrowing integer conversions, if the value is greater than the max of the target type, return the max value. If the value is less than the min of the target value then return min. float64 to float32 out of range conversions will return +-Inf. For float to int conversions we round by default but that can be modified by passing a function in the roundingMode paraneter of ConvertNumberBy
*/
func ConvertNumber[From Number, To Number](f From, t ...func(To)) To {
	return ConvertNumberBy[From, To](f)
}

func ConvertNumberBy[From Number, To Number](f From, roundMode ...func(float64) float64) To {
	var zt To
	a := &[1]To{zt}
	convertNumberBy(a, f, roundMode...)
	return a[0]
}
func convertNumberBy[From Number, To Number](a *[1]To, f From, roundMode ...func(float64) float64) {
	defer func() {
		if r := recover(); r != nil {
			a[0] = CoerceNumber[From, To](f)
		}
	}()
	var t To
	mode := math.Round
	if len(roundMode) > 0 {
		mode = roundMode[0]
	}
	isNaN := math.IsNaN(float64(f))
	istInf := math.IsInf(float64(f), 1)
	is_Inf := math.IsInf(float64(f), -1)
	z := utils.ZeroOfType[To]()
	max := MaxNum(z)
	min := MinNum(z)

	switch any(t).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		if isNaN {
			a[0] = z
			return
		}
		if istInf {
			a[0] = max
			return
		}
		if float64(f) > float64(max) {
			a[0] = max
			return
		}
		if is_Inf {
			a[0] = min
			return
		}
		if float64(f) < float64(min) {
			a[0] = min
			return
		}
		switch any(f).(type) {
		case float32, float64:
			r := mode(float64(f))
			if math.IsNaN(r) {
				a[0] = z
				return
			}
			if math.IsInf(r, 1) {
				a[0] = max
				return
			}
			if r > float64(max) {
				a[0] = max
				return
			}
			if math.IsInf(r, -1) {
				a[0] = min
				return
			}
			if r < float64(min) {
				a[0] = min
				return
			}
			a[0] = To(r)
			return
		default:
			a[0] = To(f)
			return
		}
	case float32, float64:
		a[0] = To(f)
		return
	default:
		a[0] = To(f)
		return
	}
}

/*
This is the forced number coercion function. utils.SwitchType is a function that facilitates this.
It attempts to convert the type to of the first parameter to the type passed as `func(T)` in the second parameter using a type switch.
If it fails to convert then it will do an unsafe type coercion. This is a bad idea and should be avoided but
it is the only way to get the compiler to do it sometimes.
*/
func CoerceNumber[From Number, To Number](f From, t ...func(To)) To {
	return CoerceNumberBy[From, To](f)
}
func CoerceNumberBy[From Number, To Number](f From, roundMode ...func(float64) float64) To {
	var zt To
	a := &[1]To{zt}
	coerceNumberBy(a, f, roundMode...)
	return a[0]
}
func coerceNumberBy[From Number, To Number](a *[1]To, f From, roundMode ...func(float64) float64) {
	defer func() {
		if r := recover(); r != nil {
			a[0] = utils.ZeroOfType[To]()
		}
	}()
	t := func(To) {}
	mode := math.Round
	if len(roundMode) > 0 {
		mode = roundMode[0]
	}
	isNaN := math.IsNaN(float64(f))
	istInf := math.IsInf(float64(f), 1)
	is_Inf := math.IsInf(float64(f), -1)
	z := utils.ZeroOfType[To]()
	max := MaxNum(z)
	min := MinNum(z)
	switch any(t).(type) {
	case func(int), func(int8), func(int16), func(int32), func(int64), func(uint), func(uint8), func(uint16), func(uint32), func(uint64), func(uintptr):
		if isNaN {
			a[0] = z
			return
		}
		if istInf {
			a[0] = max
			return
		}
		if float64(f) > float64(max) {
			a[0] = max
			return
		}
		if is_Inf {
			a[0] = min
			return
		}
		if float64(f) < float64(min) {
			a[0] = min
			return
		}
		switch any(f).(type) {
		case float32, float64:
			r := mode(float64(f))
			if math.IsNaN(r) {
				a[0] = z
				return
			}
			if math.IsInf(r, 1) {
				a[0] = max
				return
			}
			if r > float64(max) {
				a[0] = max
				return
			}
			if math.IsInf(r, -1) {
				a[0] = min
				return
			}
			if r < float64(min) {
				a[0] = min
				return
			}
			a[0] = To(r)
			return
		default:
			a[0] = To(f)
			return
		}
	}
	switch any(t).(type) {
	case func(int):
		a[0] = utils.SwitchType(int(f), t)
		return
	case func(int8):
		a[0] = utils.SwitchType(int8(f), t)
		return
	case func(int16):
		a[0] = utils.SwitchType(int16(f), t)
		return
	case func(int32):
		a[0] = utils.ForceType(int32(f), t)
		return
	case func(int64):
		a[0] = utils.SwitchType(int64(f), t)
		return
	case func(uint):
		a[0] = utils.SwitchType(uint(f), t)
		return
	case func(uint8):
		a[0] = utils.SwitchType(uint8(f), t)
		return
	case func(uint16):
		a[0] = utils.SwitchType(uint16(f), t)
		return
	case func(uint32):
		a[0] = utils.SwitchType(uint32(f), t)
		return
	case func(uint64):
		a[0] = utils.SwitchType(uint64(f), t)
		return
	case func(uintptr):
		a[0] = utils.SwitchType(uintptr(f), t)
		return
	case func(float32):
		a[0] = utils.SwitchType(float32(f), t)
		return
	case func(float64):
		a[0] = utils.SwitchType(float64(f), t)
		return
	default:
		a[0] = utils.ForceType(f, t)
		return
	}
}

/* This function takes in the minimum number from the bottom of the range for an individual type from the list of constants. The value is returned as a generic Number type*/
func MinNum[N Number](num ...N) N {
	var n N
	switch any(n).(type) {
	case int:
		return N(AsNumber(MinInt))
	case int8:
		return N(AsNumber(MinInt8))
	case int16:
		return N(AsNumber(MinInt16))
	case int32:
		return N(AsNumber(MinInt32))
	case int64:
		return N(AsNumber(MinInt64))
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return N(AsNumber(0))
	case float32:
		return N(AsNumber(MinFloat32))
	case float64:
		return N(AsNumber(MinFloat64))
	default:
		return utils.ZeroOfType[N]()
	}
}

/* This function takes in the maximum number from the top of the range for an individual type from the list of constants. The value is returned as a generic Number type*/
func MaxNum[N Number](num ...N) N {
	var n N
	switch any(n).(type) {
	case int:
		return N(AsNumber(MaxInt))
	case int8:
		return N(AsNumber(MaxInt8))
	case int16:
		return N(AsNumber(MaxInt16))
	case int32:
		return N(AsNumber(MaxInt32))
	case int64:
		return N(AsNumber(MaxInt64))
	case uint:
		return N(AsNumber(MaxUint))
	case uint8:
		return N(AsNumber(MaxUint8))
	case uint16:
		return N(AsNumber(MaxUint16))
	case uint32:
		return N(AsNumber(MaxUint32))
	case uint64:
		return N(AsNumber(MaxUint64))
	case uintptr:
		return N(AsNumber(MaxUintptr))
	case float32:
		return N(AsNumber(MaxFloat32))
	case float64:
		return N(AsNumber(MaxFloat64))
	default:
		return utils.ZeroOfType[N]()
	}
}

/* Here starts the functions  ̶s̶t̶o̶l̶e̶n̶taken directly from std math package. You can expect them to behave the same*/

/*
	This finds the max of a list of numbers. They can be of any number type as long as they are the same type.

The original math.Max only evaluates the max of 2 values. This maintains that functionality but is more flexible
*/
func Max[N Number](nums ...N) N {
	var max N = MinNum[N]()
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max
}

/*
	This finds the min of a list of numbers. They can be of any number type as long as they are the same type.

The original math.Min only evaluates the min of 2 values. This maintains that functionality but is more flexible
*/
func Min[N Number](nums ...N) N {
	var min N = MaxNum[N]()
	for _, num := range nums {
		if num < min {
			min = num
		}
	}
	return min
}

func Abs[N Number](num N) N {
	if num < 0 {
		return -num
	}
	return num
}

func Acos[N Number](num N) N {
	return ConvertNumber[float64, N](math.Acos(float64(num)))
}

func Acosh[N Number](num N) N {
	return ConvertNumber[float64, N](math.Acosh(float64(num)))
}

func Asin[N Number](num N) N {
	return ConvertNumber[float64, N](math.Asin(float64(num)))
}

func Asinh[N Number](num N) N {
	return ConvertNumber[float64, N](math.Asinh(float64(num)))
}

func Atan[N Number](num N) N {
	return ConvertNumber[float64, N](math.Atan(float64(num)))
}

func Atan2[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Atan2(float64(x), float64(y)))
}

func Atanh[N Number](num N) N {
	return ConvertNumber[float64, N](math.Atanh(float64(num)))
}

func Cbrt[N Number](num N) N {
	return ConvertNumber[float64, N](math.Cbrt(float64(num)))
}

func Ceil[N Number](num N) N {
	return ConvertNumberBy[float64, N](math.Ceil(float64(num)), math.Ceil)
}

func Copysign[N Number, M Number](f N, sign M) N {
	return ConvertNumber[float64, N](math.Copysign(float64(f), float64(sign)))
}

func Cos[N Number](num N) N {
	return ConvertNumber[float64, N](math.Cos(float64(num)))
}

func Cosh[N Number](num N) N {
	return ConvertNumber[float64, N](math.Cosh(float64(num)))
}

func Dim[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Dim(float64(x), float64(y)))
}

func Erf[N Number](num N) N {
	return ConvertNumber[float64, N](math.Erf(float64(num)))
}

func Erfc[N Number](num N) N {
	return ConvertNumber[float64, N](math.Erfc(float64(num)))
}

func Erfcinv[N Number](num N) N {
	return ConvertNumber[float64, N](math.Erfcinv(float64(num)))
}

func Erfinv[N Number](num N) N {
	return ConvertNumber[float64, N](math.Erfinv(float64(num)))
}

func Exp[N Number](num N) N {
	return ConvertNumber[float64, N](math.Exp(float64(num)))
}

func Exp2[N Number](num N) N {
	return ConvertNumber[float64, N](math.Exp2(float64(num)))
}

func Expm1[N Number](num N) N {
	return ConvertNumber[float64, N](math.Expm1(float64(num)))
}

func FMA[N Number, M Number, W Number](x N, y M, z W) N {
	return ConvertNumber[float64, N](math.FMA(float64(x), float64(y), float64(z)))
}

func Float32bits[N Number](x N) uint32 {
	return math.Float32bits(ConvertNumber[N, float32](x))
}

func Float32frombits[N Number](x N) float32 {
	return math.Float32frombits(ConvertNumber[N, uint32](x))
}

func Float64bits[N Number](x N) uint64 {
	return math.Float64bits(float64(x))
}

func Float64frombits[N Number](x N) float64 {
	return math.Float64frombits(ConvertNumber[N, uint64](x))
}

func Floor[N Number](num N) N {
	return ConvertNumberBy[float64, N](math.Floor(float64(num)), math.Floor)
}

func Frexp[N Number](num N) (frac float64, exp int) {
	return math.Frexp(float64(num))
}

func Gamma[N Number](num N) N {
	return ConvertNumber[float64, N](math.Gamma(float64(num)))
}

func Hypot[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Hypot(float64(x), float64(y)))
}

func Ilogb[N Number](num N) N {
	return ConvertNumber[float64, N](math.Gamma(float64(num)))
}

func Inf[N Number](num N) float64 {
	return math.Inf(ConvertNumber[N, int](num))
}

func IsInf[N Number, M Number](x N, y M) bool {
	return math.IsInf(float64(x), ConvertNumber[M, int](y))
}

func IsNaN[N Number](num N) bool {
	return math.IsNaN(float64(num))
}

func J0[N Number](num N) N {
	return ConvertNumber[float64, N](math.J0(float64(num)))
}

func J1[N Number](num N) N {
	return ConvertNumber[float64, N](math.J1(float64(num)))
}

func Jn[N Number, M Number](x N, y M) M {
	return ConvertNumber[float64, M](math.Jn(ConvertNumber[N, int](x), float64(y)))
}

func Ldexp[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Ldexp(float64(x), ConvertNumber[M, int](y)))
}

func Lgamma[N Number](num N) (N, int) {
	x, i := math.Lgamma(float64(num))
	return ConvertNumber[float64, N](x), i
}

/*I prefer at least the option for single value returns when it makes sense*/
func Lgam[N Number](num N) N {
	x, i := math.Lgamma(float64(num))
	g := ConvertNumber[float64, N](x * float64(i))
	h := ConvertNumber[float64, N](x)
	if Abs(h) > Abs(g) {
		return h
	}
	return g
}

func Log[N Number](num N) N {
	return ConvertNumber[float64, N](math.Log(float64(num)))
}

func Log10[N Number](num N) N {
	return ConvertNumber[float64, N](math.Log10(float64(num)))
}

func Log1p[N Number](num N) N {
	return ConvertNumber[float64, N](math.Log1p(float64(num)))
}

func Log2[N Number](num N) N {
	return ConvertNumber[float64, N](math.Log2(float64(num)))
}

func Logb[N Number](num N) N {
	return ConvertNumber[float64, N](math.Logb(float64(num)))
}

func Mod[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Mod(float64(x), float64(y)))
}

func Modf[N Number](num N) (N, float64) {
	x, i := math.Modf(float64(num))
	return ConvertNumber[float64, N](x), i
}

func NaN[N Number](n ...func(N)) N {
	return ConvertNumber[float64, N](math.NaN())
}

func Nextafter[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Nextafter(float64(x), float64(y)))
}

func Nextafter32[N Number, M Number](x N, y M) float32 {
	return math.Nextafter32(float32(x), float32(y))
}

func Pow[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Pow(float64(x), float64(y)))
}

func Pow10[N Number](num N) N {
	return ConvertNumber[float64, N](math.Pow10(ConvertNumber[N, int](num)))
}

func Remainder[N Number, M Number](x N, y M) N {
	return ConvertNumber[float64, N](math.Remainder(float64(x), float64(y)))
}

func Round[N Number](num N) N {
	return ConvertNumberBy[float64, N](math.Round(float64(num)), math.Round)
}

func RoundToEven[N Number](num N) N {
	return ConvertNumberBy[float64, N](math.RoundToEven(float64(num)), math.RoundToEven)
}

func Signbit[N Number](num N) bool {
	return math.Signbit(float64(num))
}

func Sin[N Number](num N) N {
	return ConvertNumber[float64, N](math.Sin(float64(num)))
}

func Sincos[N Number](num N) (N, N) {
	x, y := math.Sincos(float64(num))
	return ConvertNumber[float64, N](x), ConvertNumber[float64, N](y)
}

func Sinh[N Number](num N) N {
	return ConvertNumber[float64, N](math.Sinh(float64(num)))
}

func Sqrt[N Number](num N) N {
	return ConvertNumber[float64, N](math.Sqrt(float64(num)))
}

func Tan[N Number](num N) N {
	return ConvertNumber[float64, N](math.Tan(float64(num)))
}

func Tanh[N Number](num N) N {
	return ConvertNumber[float64, N](math.Tanh(float64(num)))
}

func Trunc[N Number](num N) N {
	return ConvertNumberBy[float64, N](math.Trunc(float64(num)), math.Trunc)
}

func Y0[N Number](num N) N {
	return ConvertNumber[float64, N](math.Y0(float64(num)))
}

func Y1[N Number](num N) N {
	return ConvertNumber[float64, N](math.Y1(float64(num)))
}

func Yn[N Number, M Number](x N, y M) M {
	return ConvertNumber[float64, M](math.Yn(ConvertNumber[N, int](x), float64(y)))
}
