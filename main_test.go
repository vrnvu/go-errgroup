package main

import (
	"fmt"
	"testing"
)

func assert(t *testing.T, got []int, want []int) {
	if len(got) != len(want) {
		t.Fatalf("len got != len want")
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%d != %d", got[i], want[i])
		}
	}

}

// returns a slice from 0 to N
// [0, N)
func buildInput(N int) []int {
	r := make([]int, N)
	for i := 0; i < N; i++ {
		r[i] = i
	}
	return r
}

func buildOutput(input []int) []int {
	r := make([]int, len(input))
	for i, v := range input {
		r[i] = v * 2
	}
	return r
}

func TestWorkers(t *testing.T) {
	for inputSize := 10; inputSize <= 10000; inputSize = inputSize * 10 {
		input := buildInput(inputSize)
		want := buildOutput(input)
		for workers := 1; workers < 10; workers++ {
			t.Run(fmt.Sprintf("inputSize: %d workers %d", inputSize, workers), func(t *testing.T) {
				got := Workers(workers, input)
				assert(t, got, want)
			})
		}
	}
}
