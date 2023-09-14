package tests

import (
	"fmt"
	"testing"
)

func alignFloor(value, base int) int {
	return ((value) & ^((base) - 1))
}
func alignCeil(value, base int) int {
	return alignFloor((value)+((base)-1), base)
}

func alignRound(value, base int) int {
	return alignFloor((value)+((base)/2), base)
}

func TestAlign(t *testing.T) {
	fmt.Println(alignFloor(31, 16))
	fmt.Println(alignCeil(31, 16))
	fmt.Println(alignRound(31, 16))

	fmt.Println(alignFloor(32, 16))
	fmt.Println(alignCeil(32, 16))
	fmt.Println(alignRound(32, 16))

	fmt.Println(alignFloor(33, 16))
	fmt.Println(alignCeil(33, 16))
	fmt.Println(alignRound(33, 16))
}
