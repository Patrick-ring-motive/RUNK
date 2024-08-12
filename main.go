package main

//package Maths
import (
	"fmt"
	//	"math"
	. "github.com/Patrick-ring-motive/utils"
	"strconv"
)

const (
	MaxInt                 = 1<<(strconv.IntSize-1) - 1    // MaxInt32 or MaxInt64 depending on intSize.
	MinInt                 = -1 << (strconv.IntSize - 1)   // MinInt32 or MinInt64 depending on intSize.
	MaxInt8                = 1<<7 - 1                      // 127
	MinInt8                = -1 << 7                       // -128
	MaxInt16               = 1<<15 - 1                     // 32767
	MinInt16               = -1 << 15                      // -32768
	MaxInt32               = 1<<31 - 1                     // 2147483647
	MinInt32               = -1 << 31                      // -2147483648
	MaxInt64               = 1<<63 - 1                     // 9223372036854775807
	MinInt64               = -1 << 63                      // -9223372036854775808
	MaxUint                = 1<<strconv.IntSize - 1        // MaxUint32 or MaxUint64 depending on intSize.
	MaxUint8               = 1<<8 - 1                      // 255
	MaxUint16              = 1<<16 - 1                     // 65535
	MaxUint32              = 1<<32 - 1                     // 4294967295
	MaxUint64              = 1<<64 - 1                     // 18446744073709551615
	MaxFloat32             = 0x1p127 * (1 + (1 - 0x1p-23)) // 3.40282346638528859811704183484516925440e+38
	MinFloat32             = -MaxFloat32
	SmallestNonzeroFloat32 = 0x1p-126 * 0x1p-23 // 1.401298464324817070923729583289916131280e-45

	MaxFloat64             = 0x1p1023 * (1 + (1 - 0x1p-52)) // 1.79769313486231570814527423731704356798070e+308
	MinFloat64             = -MaxFloat64
	SmallestNonzeroFloat64 = 0x1p-1022 * 0x1p-52 // 4.9406564584124654417656879286822137236505980e-324
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func MinNum[N Number](num N) N {
	switch any(num).(type) {
	case int:
		return SwitchType(MinInt, func(n N) {})
	case int8:
		return SwitchType(MinInt8, func(n N) {})
	case int16:
		return SwitchType(MinInt16, func(n N) {})
	case int32:
		return SwitchType(MinInt32, func(n N) {})
	case int64:
		return SwitchType(MinInt64, func(n N) {})
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return SwitchType(0, func(n N) {})
	case float32:
		return SwitchType(MinFloat32, func(n N) {})
	case float64:
		return SwitchType(MinFloat64, func(n N) {})
	default:
		return ZeroOfType(func(n N) {})
	}
}

func MaxNum[N Number](num N) N {
	switch any(num).(type) {
	case int:
		return SwitchType(MaxInt, func(n N) {})
	case int8:
		return SwitchType(MaxInt8, func(n N) {})
	case int16:
		return SwitchType(MaxInt16, func(n N) {})
	case int32:
		return SwitchType(MaxInt32, func(n N) {})
	case int64:
		return SwitchType(MaxInt64, func(n N) {})
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return SwitchType(0, func(n N) {})
	case float32:
		return SwitchType(MaxFloat32, func(n N) {})
	case float64:
		return SwitchType(MaxFloat64, func(n N) {})
	default:
		return ZeroOfType(func(n N) {})
	}
}

func Max[N Number](nums ...N) N {
	var max N = MinNum(nums[0])
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max
}

func Min[N Number](nums ...N) N {
	var min N = MinNum(nums[0])
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

func main() {
	fmt.Println(Abs(-0.5))
}
