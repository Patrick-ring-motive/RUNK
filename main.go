package main

//package maths
/*
The maths package provides generic alternatives to the std math package to allow usage
with any of go's number types. There are some minor additions added for convenience but the original
behavior is preserved.
*/
import (
	"fmt"
	"github.com/Patrick-ring-motive/utils"
	"math"
	"strconv"
)

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
	SmallestNonzeroFloat32 float32 = 0x1p-126 * 0x1p-23 // 1.401298464324817070923729583289916131280e-45
	MaxFloat64             float64 = 0x1p1023 * (1 + (1 - 0x1p-52)) // 1.79769313486231570814527423731704356798070e+308
	MinFloat64             float64 = -MaxFloat64
	SmallestNonzeroFloat64 float64 = 0x1p-1022 * 0x1p-52 // 4.9406564584124654417656879286822137236505980e-324
)

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


/*This is the helper function that makes all this possible without having to write a different implementation
for each indiviual type. ConvertNumber handles converting between number types while maintaining the generic
Number designation on the return. You'll see me using the `func(T)` syntax which is a pattern that I use to
pass around a type without having to instantiate it. This works to give the compiler a hint that it can use to
maitain type safety. This should work for most scenarios.
The main edge cases to worry about are NaN and Inf which can get coerced into a number that isn't very meaningful. NaN converted to an int will return 0 so that at least it maintains the same truthiness and +/- Inf converted to an int will return MaxInt/MinInt. In narrowing integer conversions, if the value is greater than the max of the target type, return the max value. If the value is less than the min of the target value then return min. float64 to float32 out of range conversions will return +-Inf. For float to int conversions we truncate*/
func ConvertNumber[From Number, To Number](f From, t func(To)) To {
	isNaN := math.IsNaN(float64(f))
	istInf := math.IsInf(float64(f), 1)
	is_Inf := math.IsInf(float64(f), -1)
	z := utils.ZeroOfType(t)
	max := MaxNum(z)
	min := MinNum(z)

	switch any(t).(type) {
	case func(int),func(int8),func(int16),func(int32),func(int64),func(uint),func(uint8),func(uint16),func(uint32),func(uint64),func(uintptr):
		if(isNaN){
			return z
		}
		if(istInf){
			return max
		}
		if(float64(f) > float64(max)){
			return max
		}
		if(is_Inf){
			return min
		}
		if(float64(f) < float64(min)){
			return min
		}
		switch any(f).(type) {
			case float32,float64:
				return To(math.Trunc(float64(f)))
			default:
				return To(f)
		}
	case func(float32),func(float64):
		return To(f)
	default:
		return To(f)
	}
}

/*
This is the forced number coercion function. utils.SwitchType is a function that facilitates this. 
It attempts to convert the type to of the first parameter to the type passed as `func(T)` in the second parameter using a type switch.
If it fails to convert then it will do an unsafe type coercion. This is a bad idea and should be avoided but
it is the only way to get the compiler to do it sometimes.
*/
func CoerceNumber[From Number, To Number](f From, t func(To)) To {
	isNaN := math.IsNaN(float64(f))
	istInf := math.IsInf(float64(f), 1)
	is_Inf := math.IsInf(float64(f), -1)
	z := utils.ZeroOfType(t)
	max := MaxNum(z)
	min := MinNum(z)
	switch any(t).(type) {
		case func(int),func(int8),func(int16),func(int32),func(int64),func(uint),func(uint8),func(uint16),func(uint32),func(uint64),func(uintptr):
		if(isNaN){
			return zt
		}
		if(istInf){
			return max
		}
		if(float64(f) > float64(max)){
			return max
		}
		if(is_Inf){
			return min
		}
		if(float64(f) < float64(min)){
			return min
		}
		switch any(f).(type) {
			case float32,float64:
				return To(math.Trunc(float64(f)))
			default:
				return To(f)
		}
	}
	switch any(t).(type) {
	case func(int):
		return utils.SwitchType(int(f), t)
	case func(int8):
		return utils.SwitchType(int8(f), t)
	case func(int16):
		return utils.SwitchType(int16(f), t)
	case func(int32):
		return utils.ForceType(int32(f), t)
	case func(int64):
		return utils.SwitchType(int64(f), t)
	case func(uint):
		return utils.SwitchType(uint(f), t)
	case func(uint8):
		return utils.SwitchType(uint8(f), t)
	case func(uint16):
		return utils.SwitchType(uint16(f), t)
	case func(uint32):
		return utils.SwitchType(uint32(f), t)
	case func(uint64):
		return utils.SwitchType(uint64(f), t)
	case func(uintptr):
		return utils.SwitchType(uintptr(f), t)
	case func(float32):
		return utils.SwitchType(float32(f), t)
	case func(float64):
		return utils.SwitchType(float64(f), t)
	default:
		return utils.ForceType(f, t)
	}
}

/* This function takes in the minimum number from the bottom of the range for an individual type from the list of constants. The value is returned as a generic Number type*/
func MinNum[N Number](num N) N {
	switch any(num).(type) {
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
		return utils.ZeroOfType(utils.TypeOf(num))
	}
}

/* This function takes in the maximum number from the top of the range for an individual type from the list of constants. The value is returned as a generic Number type*/
func MaxNum[N Number](num ...N) N {
	switch any(num[0]).(type) {
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
		return utils.ZeroOfType(utils.TypeOf(num[0]))
	}
}

/* This finds the max of a list of numbers. They can be of any number type as long as they are the same type.
The original math.Max only evaluates the max of 2 values. This maintains that functionality but is more flexible*/
func Max[N Number](nums ...N) N {
	var max N = MinNum(nums[0])
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max
}

/* This finds the min of a list of numbers. They can be of any number type as long as they are the same type.
The original math.Min only evaluates the min of 2 values. This maintains that functionality but is more flexible*/
func Min[N Number](nums ...N) N {
	var min N = MaxNum(nums[0])
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
	return ConvertNumber(math.Acos(float64(num)),utils.TypeOf(num))
}


func main() {
	fmt.Println(Acos(int32(10)))
	fmt.Println(MaxFloat64,ConvertNumber(MaxFloat64,func(i float32){}))
}
