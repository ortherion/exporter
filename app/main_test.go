package main

import (
	"fmt"
	"testing"
)

var convertTest = []struct{
	in string
	want float64
}{
	{"AA", 0xAA},
	{"12", 12},
	{"SOME", 0},
	{"12.5", 12.5},
}

func TestConvertToString(t *testing.T) {
	for _, val := range convertTest {
		name := fmt.Sprintf("case(%s,%v)", val.in, val.want)
		t.Run(name, func(t *testing.T) {
			got := ConvertToString(val.in)
			if got != val.want {
				t.Errorf("got %f; want %v", got, val.want)
			}
		})
	}
}

