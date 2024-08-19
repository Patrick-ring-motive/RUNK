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
	"reflect"
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

/* All the integers */
type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

/* All the unsigned integers */
type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

/* All the signed integers */
type Sint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

/* The floats */
type Float interface {
	~float32 | ~float64
}

/*
AsNumber is just used as a compiler hint to tell the compiler that the value can be any Number type.
This is used to make the compiler happy when using narrowing conversions.
*/
func AsNumber[N Number](n N) N {
	return n
}

/*This is the lazy built in number conversion method. Sometimes used as a fallback method.*/
func ToNumber[To Number, From Number](n From) To {
	return To(n)
}

func fromNumber[To Number, From Number](from From, roundMode ...func(float64) float64) To {
	switch any(from).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return fromInt[To](from)
	case float32, float64:
		return fromFloat[To](from, roundMode...)
	}
	v := reflect.ValueOf(from)
	kind := v.Kind()
	switch kind {
	case reflect.Float32, reflect.Float64:
		if v.CanFloat() {
			return fromFloat[To](float64(from), roundMode...)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			return fromSint[To](int64(from))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			return fromUint[To](uint64(from))
		}
	}
	return To(from)
}

func fromInt[To Number, From Number](from From) To {
	switch any(from).(type) {
	case int, int8, int16, int32, int64:
		return fromSint[To](from)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return fromUint[To](from)
	}
	v := reflect.ValueOf(from)
	kind := v.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			return fromSint[To](int64(from))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			return fromUint[To](uint64(from))
		}
	}
	return To(from)
}

func fromSint[To Number, From Number](from From) To {
	var to To
	switch any(to).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return sintToInt(from, to)
	case float32, float64:
		return sintToFloat[To](from)
	}
	v := reflect.ValueOf(to)
	kind := v.Kind()
	switch kind {
	case reflect.Float32, reflect.Float64:
		if v.CanFloat() {
			switch kind {
			case reflect.Float32:
				return To(sintToFloat[float32](from))
			case reflect.Float64:
				return To(sintToFloat[float64](from))
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			switch kind {
			case reflect.Int:
				return To(sintToSint[int](from))
			case reflect.Int8:
				return To(sintToSint[int8](from))
			case reflect.Int16:
				return To(sintToSint[int16](from))
			case reflect.Int32:
				return To(sintToSint[int32](from))
			case reflect.Int64:
				return To(sintToSint[int64](from))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			switch kind {
			case reflect.Uint:
				return To(sintToUint[uint](from))
			case reflect.Uint8:
				return To(sintToUint[uint8](from))
			case reflect.Int16:
				return To(sintToUint[uint16](from))
			case reflect.Uint32:
				return To(sintToUint[uint32](from))
			case reflect.Uint64:
				return To(sintToUint[uint64](from))
			case reflect.Uintptr:
				return To(sintToUint[uintptr](from))
			}
		}
	}
	return To(from)
}

func sintToInt[To Number, From Number](from From, to To) To {
	switch any(to).(type) {
	case int, int8, int16, int32, int64:
		return To(sintToSint[To](from))
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return To(sintToUint[To](from))
	}
	v := reflect.ValueOf(to)
	kind := v.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			switch kind {
			case reflect.Int:
				return To(sintToSint[int](from))
			case reflect.Int8:
				return To(sintToSint[int8](from))
			case reflect.Int16:
				return To(sintToSint[int16](from))
			case reflect.Int32:
				return To(sintToSint[int32](from))
			case reflect.Int64:
				return To(sintToSint[int64](from))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			switch kind {
			case reflect.Uint:
				return To(sintToUint[uint](from))
			case reflect.Uint8:
				return To(sintToUint[uint8](from))
			case reflect.Int16:
				return To(sintToUint[uint16](from))
			case reflect.Uint32:
				return To(sintToUint[uint32](from))
			case reflect.Uint64:
				return To(sintToUint[uint64](from))
			case reflect.Uintptr:
				return To(sintToUint[uintptr](from))
			}
		}
	}
	return To(from)
}

func sintToSint[To Number, From Number](from From) To {
	max := MaxNum[To]()
	sint := int64(from)
	if sint > int64(max) {
		return max
	}
	min := MinNum[To]()
	if sint < int64(min) {
		return min
	}
	return To(from)
}

func sintToUint[To Number, From Number](from From) To {
	if from < 0 {
		return To(0)
	}
	max := MaxNum[To]()
	if uint64(from) > uint64(max) {
		return max
	}
	return To(from)
}

func sintToFloat[To Number, From Number](from From) To {
	return To(from)
}

func fromUint[To Number, From Number](from From) To {
	var to To
	switch any(to).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return uintToInt(from, to)
	case float32, float64:
		return uintToFloat[To](from)
	}
	v := reflect.ValueOf(to)
	kind := v.Kind()
	switch kind {
	case reflect.Float32, reflect.Float64:
		if v.CanFloat() {
			switch kind {
			case reflect.Float32:
				return To(uintToFloat[float32](from))
			case reflect.Float64:
				return To(uintToFloat[float64](from))
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			switch kind {
			case reflect.Int:
				return To(uintToSint[int](from))
			case reflect.Int8:
				return To(uintToSint[int8](from))
			case reflect.Int16:
				return To(uintToSint[int16](from))
			case reflect.Int32:
				return To(uintToSint[int32](from))
			case reflect.Int64:
				return To(uintToSint[int64](from))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			switch kind {
			case reflect.Uint:
				return To(uintToUint[uint](from))
			case reflect.Uint8:
				return To(uintToUint[uint8](from))
			case reflect.Int16:
				return To(uintToUint[uint16](from))
			case reflect.Uint32:
				return To(uintToUint[uint32](from))
			case reflect.Uint64:
				return To(uintToUint[uint64](from))
			case reflect.Uintptr:
				return To(uintToUint[uintptr](from))
			}
		}
	}
	return To(from)
}

func uintToInt[To Number, From Number](from From, to To) To {
	switch any(to).(type) {
	case int, int8, int16, int32, int64:
		return To(uintToSint[To](from))
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return To(uintToUint[To](from))
	}
	v := reflect.ValueOf(to)
	kind := v.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			switch kind {
			case reflect.Int:
				return To(uintToSint[int](from))
			case reflect.Int8:
				return To(uintToSint[int8](from))
			case reflect.Int16:
				return To(uintToSint[int16](from))
			case reflect.Int32:
				return To(uintToSint[int32](from))
			case reflect.Int64:
				return To(uintToSint[int64](from))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			switch kind {
			case reflect.Uint:
				return To(uintToUint[uint](from))
			case reflect.Uint8:
				return To(uintToUint[uint8](from))
			case reflect.Int16:
				return To(uintToUint[uint16](from))
			case reflect.Uint32:
				return To(uintToUint[uint32](from))
			case reflect.Uint64:
				return To(uintToUint[uint64](from))
			case reflect.Uintptr:
				return To(uintToUint[uintptr](from))
			}
		}
	}
	return To(from)
}

func uintToSint[To Number, From Number](from From) To {
	max := MaxNum[To]()
	if uint64(from) > uint64(max) {
		return max
	}
	return To(from)
}

func uintToUint[To Number, From Number](from From) To {
	max := MaxNum[To]()
	if uint64(from) > uint64(max) {
		return max
	}
	return To(from)
}

func uintToFloat[To Number, From Number](from From) To {
	return To(from)
}

func fromFloat[To Number, From Number](from From, roundMode ...func(float64) float64) To {
	var to To
	switch any(to).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		mode := math.Round
		if len(roundMode) > 0 {
			mode = roundMode[0]
		}
		return floatToInt(from, to, mode)
	case float32, float64:
		return floatToFloat[To](from)
	}
	v := reflect.ValueOf(to)
	kind := v.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			mode := math.Round
			if len(roundMode) > 0 {
				mode = roundMode[0]
			}
			switch kind {
			case reflect.Int:
				return To(floatToSint[int](from, mode))
			case reflect.Int8:
				return To(floatToSint[int8](from, mode))
			case reflect.Int16:
				return To(floatToSint[int16](from, mode))
			case reflect.Int32:
				return To(floatToSint[int32](from, mode))
			case reflect.Int64:
				return To(floatToSint[int64](from, mode))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			mode := math.Round
			if len(roundMode) > 0 {
				mode = roundMode[0]
			}
			switch kind {
			case reflect.Uint:
				return To(floatToUint[uint](from, mode))
			case reflect.Uint8:
				return To(floatToUint[uint8](from, mode))
			case reflect.Int16:
				return To(floatToUint[uint16](from, mode))
			case reflect.Uint32:
				return To(floatToUint[uint32](from, mode))
			case reflect.Uint64:
				return To(floatToUint[uint64](from, mode))
			case reflect.Uintptr:
				return To(floatToUint[uintptr](from, mode))
			}
		}
	}
	return To(from)
}

func floatToInt[To Number, From Number](from From, to To, roundMode func(float64) float64) To {
	switch any(to).(type) {
	case int, int8, int16, int32, int64:
		return floatToSint[To](from, roundMode)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return floatToUint[To](from, roundMode)
	}
	v := reflect.ValueOf(to)
	kind := v.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			switch kind {
			case reflect.Int:
				return To(floatToSint[int](from, roundMode))
			case reflect.Int8:
				return To(floatToSint[int8](from, roundMode))
			case reflect.Int16:
				return To(floatToSint[int16](from, roundMode))
			case reflect.Int32:
				return To(floatToSint[int32](from, roundMode))
			case reflect.Int64:
				return To(floatToSint[int64](from, roundMode))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			switch kind {
			case reflect.Uint:
				return To(floatToUint[uint](from, roundMode))
			case reflect.Uint8:
				return To(floatToUint[uint8](from, roundMode))
			case reflect.Int16:
				return To(floatToUint[uint16](from, roundMode))
			case reflect.Uint32:
				return To(floatToUint[uint32](from, roundMode))
			case reflect.Uint64:
				return To(floatToUint[uint64](from, roundMode))
			case reflect.Uintptr:
				return To(floatToUint[uintptr](from, roundMode))
			}
		}
	}
	return To(from)
}

func floatToSint[To Number, From Number](from From, roundMode func(float64) float64) To {
	fl := float64(from)
	r := roundMode(fl)
	if math.IsNaN(fl) {
		return To(0)
	}
	max := MaxNum[To]()
	flmax := float64(max)
	if math.IsInf(fl, 1) || fl > flmax || r > flmax {
		return max
	}
	min := MinNum[To]()
	flmin := float64(min)
	if math.IsInf(fl, -1) || fl < flmin || r < flmin {
		return min
	}
	return To(r)
}

func floatToUint[To Number, From Number](from From, roundMode func(float64) float64) To {
	fl := float64(from)
	r := roundMode(fl)
	if math.IsNaN(fl) || math.IsInf(fl, -1) || fl < 0.0 || r < 0.0 {
		return To(0)
	}
	max := MaxNum[To]()
	flmax := float64(max)
	if math.IsInf(fl, 1) || fl > flmax || r > flmax {
		return max
	}
	return To(from)
}

func floatToFloat[To Number, From Number](from From) To {
	return To(from)
}

/*
ConvertNum is the most flexible conversion function as it accepts an any type.
It is needed to make number conversion more concise. Even though it is intended to use with numbers, it will make a best effort to convert non number types. Typical usade looks like
`ConvertNum[int](11.2)` which will return 11.
*/
func ConvertNum[To Number](f any) To {
	if f == nil {
		return To(0)
	}
	switch v := f.(type) {
	case int:
		return ConvertNumber[To](v)
	case int8:
		return ConvertNumber[To](v)
	case int16:
		return ConvertNumber[To](v)
	case int32:
		return ConvertNumber[To](v)
	case int64:
		return ConvertNumber[To](v)
	case uint:
		ConvertNumber[To](v)
	case uint8:
		return ConvertNumber[To](v)
	case uint16:
		return ConvertNumber[To](v)
	case uint32:
		return ConvertNumber[To](v)
	case uint64:
		return ConvertNumber[To](v)
	case uintptr:
		return ConvertNumber[To](v)
	case float32:
		return ConvertNumber[To](v)
	case float64:
		return ConvertNumber[To](v)
	case bool:
		if v {
			return To(1)
		} else {
			return To(0)
		}
	default:
		switch k := reflect.ValueOf(f); k.Kind() {
		case reflect.Float32, reflect.Float64:
			if k.CanFloat() {
				return ConvertNumber[To](k.Float())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if k.CanInt() {
				return ConvertNumber[To](k.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if k.CanUint() {
				return ConvertNumber[To](k.Uint())
			}
		}
	}
	return utils.Convert[To](f)
}

/*
This is the helper function that makes all this possible without having to write a different implementation
for each indiviual type. ConvertNumber handles converting between number types while maintaining the generic
Number designation on the return. You'll see me using the `func(T)` syntax which is a pattern that I use to
pass around a type without having to instantiate it. This works to give the compiler a hint that it can use to
maitain type safety. This should work for most scenarios.
The main edge cases to worry about are NaN and Inf which can get coerced into a number that isn't very meaningful. NaN converted to an int will return 0 so that at least it maintains the same truthiness and +/- Inf converted to an int will return MaxInt/MinInt. In narrowing integer conversions, if the value is greater than the max of the target type, return the max value. If the value is less than the min of the target value then return min. float64 to float32 out of range conversions will return +-Inf. For float To Number conversions we round by default but that can be modified by passing a function in the roundingMode paraneter of ConvertNumberBy
*/
func ConvertNumber[To Number, From Number](f From) To {
	return ConvertNumberBy[To](f)
}

func ConvertNumberBy[To Number, From Number](from From, roundMode ...func(float64) float64) To {
	var zt To
	a := &[1]To{zt}
	convertNumberBy[To](a, from, roundMode...)
	return a[0]
}
func convertNumberBy[To Number, From Number](a *[1]To, from From, roundMode ...func(float64) float64) {
	defer func() {
		if r := recover(); r != nil {
			a[0] = To(from)
		}
	}()
	mode := math.Round
	if len(roundMode) > 0 {
		mode = roundMode[0]
	}
	switch any(from).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		a[0] = fromNumber[To](from, roundMode...)
		return
	}
	v := reflect.ValueOf(from)
	kind := v.Kind()
	switch kind {
	case reflect.Float32, reflect.Float64:
		if v.CanFloat() {
			a[0] = fromFloat[To](float64(from), mode)
			return
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanInt() {
			a[0] = fromSint[To](int64(from))
			return
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.CanUint() {
			a[0] = fromUint[To](uint64(from))
			return
		}
	}
	a[0] = To(from)
}

/* This function takes in the minimum number from the bottom of the range for an individual type from the list of constants. The value is returned as a generic Number type*/
func MinNum[Num Number](num ...Num) Num {
	var n Num
	switch any(n).(type) {
	case int:
		return ToNumber[Num](MinInt)
	case int8:
		return ToNumber[Num](MinInt8)
	case int16:
		return ToNumber[Num](MinInt16)
	case int32:
		return ToNumber[Num](MinInt32)
	case int64:
		return ToNumber[Num](MinInt64)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return ToNumber[Num](0)
	case float32:
		return ToNumber[Num](MinFloat32)
	case float64:
		return ToNumber[Num](MinFloat64)
	}
	switch k := reflect.ValueOf(n); k.Kind() {
	case reflect.Int:
		return ToNumber[Num](MinInt)
	case reflect.Int8:
		return ToNumber[Num](MinInt8)
	case reflect.Int16:
		return ToNumber[Num](MinInt16)
	case reflect.Int32:
		return ToNumber[Num](MinInt32)
	case reflect.Int64:
		return ToNumber[Num](MinInt64)
	case reflect.Float32:
		return ToNumber[Num](MinFloat32)
	case reflect.Float64:
		return ToNumber[Num](MinFloat64)
	}
	return n
}

/* This function takes in the maximum number from the top of the range for an individual type from the list of constants. The value is returned as a generic Number type*/
func MaxNum[Num Number](num ...Num) Num {
	var n Num
	switch any(n).(type) {
	case int:
		return ToNumber[Num](MaxInt)
	case int8:
		return ToNumber[Num](MaxInt8)
	case int16:
		return ToNumber[Num](MaxInt16)
	case int32:
		return ToNumber[Num](MaxInt32)
	case int64:
		return ToNumber[Num](MaxInt64)
	case uint:
		return ToNumber[Num](MaxUint)
	case uint8:
		return ToNumber[Num](MaxUint8)
	case uint16:
		return ToNumber[Num](MaxUint16)
	case uint32:
		return ToNumber[Num](MaxUint32)
	case uint64:
		return ToNumber[Num](MaxUint64)
	case uintptr:
		return ToNumber[Num](MaxUintptr)
	case float32:
		return ToNumber[Num](MaxFloat32)
	case float64:
		return ToNumber[Num](MaxFloat64)
	}
	switch k := reflect.ValueOf(n); k.Kind() {
	case reflect.Int:
		return ToNumber[Num](MaxInt)
	case reflect.Int8:
		return ToNumber[Num](MaxInt8)
	case reflect.Int16:
		return ToNumber[Num](MaxInt16)
	case reflect.Int32:
		return ToNumber[Num](MaxInt32)
	case reflect.Int64:
		return ToNumber[Num](MaxInt64)
	case reflect.Uint:
		return ToNumber[Num](MaxUint)
	case reflect.Uint8:
		return ToNumber[Num](MaxUint8)
	case reflect.Uint16:
		return ToNumber[Num](MaxUint16)
	case reflect.Uint32:
		return ToNumber[Num](MaxUint32)
	case reflect.Uint64:
		return ToNumber[Num](MaxUint64)
	case reflect.Uintptr:
		return ToNumber[Num](MaxUintptr)
	case reflect.Float32:
		return ToNumber[Num](MaxFloat32)
	case reflect.Float64:
		return ToNumber[Num](MaxFloat64)
	}
	return n
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
	return ConvertNumber[N](math.Acos(float64(num)))
}

func Acosh[N Number](num N) N {
	return ConvertNumber[N](math.Acosh(float64(num)))
}

func Asin[N Number](num N) N {
	return ConvertNumber[N](math.Asin(float64(num)))
}

func Asinh[N Number](num N) N {
	return ConvertNumber[N](math.Asinh(float64(num)))
}

func Atan[N Number](num N) N {
	return ConvertNumber[N](math.Atan(float64(num)))
}

func Atan2[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Atan2(float64(x), float64(y)))
}

func Atanh[N Number](num N) N {
	return ConvertNumber[N](math.Atanh(float64(num)))
}

func Cbrt[N Number](num N) N {
	return ConvertNumber[N](math.Cbrt(float64(num)))
}

func Ceil[N Number](num N) N {
	return ConvertNumberBy[N](math.Ceil(float64(num)), math.Ceil)
}

func Copysign[N Number, M Number](f N, sign M) N {
	return ConvertNumber[N](math.Copysign(float64(f), float64(sign)))
}

func Cos[N Number](num N) N {
	return ConvertNumber[N](math.Cos(float64(num)))
}

func Cosh[N Number](num N) N {
	return ConvertNumber[N](math.Cosh(float64(num)))
}

func Dim[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Dim(float64(x), float64(y)))
}

func Erf[N Number](num N) N {
	return ConvertNumber[N](math.Erf(float64(num)))
}

func Erfc[N Number](num N) N {
	return ConvertNumber[N](math.Erfc(float64(num)))
}

func Erfcinv[N Number](num N) N {
	return ConvertNumber[N](math.Erfcinv(float64(num)))
}

func Erfinv[N Number](num N) N {
	return ConvertNumber[N](math.Erfinv(float64(num)))
}

func Exp[N Number](num N) N {
	return ConvertNumber[N](math.Exp(float64(num)))
}

func Exp2[N Number](num N) N {
	return ConvertNumber[N](math.Exp2(float64(num)))
}

func Expm1[N Number](num N) N {
	return ConvertNumber[N](math.Expm1(float64(num)))
}

func FMA[N Number, M Number, W Number](x N, y M, z W) N {
	return ConvertNumber[N](math.FMA(float64(x), float64(y), float64(z)))
}

func Float32bits[N Number](x N) uint32 {
	return math.Float32bits(ConvertNumber[float32](x))
}

func Float32frombits[N Number](x N) float32 {
	return math.Float32frombits(ConvertNumber[uint32](x))
}

func Float64bits[N Number](x N) uint64 {
	return math.Float64bits(float64(x))
}

func Float64frombits[N Number](x N) float64 {
	return math.Float64frombits(ConvertNumber[uint64](x))
}

func Floor[N Number](num N) N {
	return ConvertNumberBy[N](math.Floor(float64(num)), math.Floor)
}

func Frexp[N Number](num N) (frac float64, exp int) {
	return math.Frexp(float64(num))
}

func Gamma[N Number](num N) N {
	return ConvertNumber[N](math.Gamma(float64(num)))
}

func Hypot[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Hypot(float64(x), float64(y)))
}

func Ilogb[N Number](num N) N {
	return ConvertNumber[N](math.Gamma(float64(num)))
}

func Inf[N Number](num N) float64 {
	return math.Inf(ConvertNumber[int](num))
}

func IsInf[N Number, M Number](x N, y M) bool {
	return math.IsInf(float64(x), ConvertNumber[int](y))
}

func IsNaN[N Number](num N) bool {
	return math.IsNaN(float64(num))
}

func J0[N Number](num N) N {
	return ConvertNumber[N](math.J0(float64(num)))
}

func J1[N Number](num N) N {
	return ConvertNumber[N](math.J1(float64(num)))
}

func Jn[N Number, M Number](x N, y M) M {
	return ConvertNumber[M](math.Jn(ConvertNumber[int](x), float64(y)))
}

func Ldexp[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Ldexp(float64(x), ConvertNumber[int](y)))
}

func Lgamma[N Number](num N) (N, int) {
	x, i := math.Lgamma(float64(num))
	return ConvertNumber[N](x), i
}

/*I prefer at least the option for single value returns when it makes sense*/
func Lgam[N Number](num N) N {
	x, i := math.Lgamma(float64(num))
	g := ConvertNumber[N](x * float64(i))
	h := ConvertNumber[N](x)
	if Abs(h) > Abs(g) {
		return h
	}
	return g
}

func Log[N Number](num N) N {
	return ConvertNumber[N](math.Log(float64(num)))
}

func Log10[N Number](num N) N {
	return ConvertNumber[N](math.Log10(float64(num)))
}

func Log1p[N Number](num N) N {
	return ConvertNumber[N](math.Log1p(float64(num)))
}

func Log2[N Number](num N) N {
	return ConvertNumber[N](math.Log2(float64(num)))
}

func Logb[N Number](num N) N {
	return ConvertNumber[N](math.Logb(float64(num)))
}

func Mod[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Mod(float64(x), float64(y)))
}

func Modf[N Number](num N) (N, float64) {
	x, i := math.Modf(float64(num))
	return ConvertNumber[N](x), i
}

func NaN[N Number](n ...func(N)) N {
	return ConvertNumber[N](math.NaN())
}

func Nextafter[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Nextafter(float64(x), float64(y)))
}

func Nextafter32[N Number, M Number](x N, y M) float32 {
	return math.Nextafter32(float32(x), float32(y))
}

func Pow[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Pow(float64(x), float64(y)))
}

func Pow10[N Number](num N) N {
	return ConvertNumber[N](math.Pow10(ConvertNumber[int](num)))
}

func Remainder[N Number, M Number](x N, y M) N {
	return ConvertNumber[N](math.Remainder(float64(x), float64(y)))
}

func Round[N Number](num N) N {
	return ConvertNumberBy[N](math.Round(float64(num)), math.Round)
}

func RoundToEven[N Number](num N) N {
	return ConvertNumberBy[N](math.RoundToEven(float64(num)), math.RoundToEven)
}

func Signbit[N Number](num N) bool {
	return math.Signbit(float64(num))
}

func Sin[N Number](num N) N {
	return ConvertNumber[N](math.Sin(float64(num)))
}

func Sincos[N Number](num N) (N, N) {
	x, y := math.Sincos(float64(num))
	return ConvertNumber[N](x), ConvertNumber[N](y)
}

func Sinh[N Number](num N) N {
	return ConvertNumber[N](math.Sinh(float64(num)))
}

func Sqrt[N Number](num N) N {
	return ConvertNumber[N](math.Sqrt(float64(num)))
}

func Tan[N Number](num N) N {
	return ConvertNumber[N](math.Tan(float64(num)))
}

func Tanh[N Number](num N) N {
	return ConvertNumber[N](math.Tanh(float64(num)))
}

func Trunc[N Number](num N) N {
	return ConvertNumberBy[N](math.Trunc(float64(num)), math.Trunc)
}

func Y0[N Number](num N) N {
	return ConvertNumber[N](math.Y0(float64(num)))
}

func Y1[N Number](num N) N {
	return ConvertNumber[N](math.Y1(float64(num)))
}

func Yn[N Number, M Number](x N, y M) M {
	return ConvertNumber[M](math.Yn(ConvertNumber[int](x), float64(y)))
}
