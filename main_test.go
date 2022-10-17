package main

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/sync/errgroup"
)

func assert(t *testing.T, got []int, want []int) {
	if len(got) != len(want) {
		t.Fatalf("len got != len want, %v = %v", got, want)
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

func TestBackground(t *testing.T) {
	for inputSize := 10; inputSize <= 10; inputSize = inputSize + 100 {
		input := buildInput(inputSize)
		want := buildOutput(input)
		for workers := 6; workers <= 6; workers++ {
			s := newService(workers)
			t.Run(fmt.Sprintf("inputSize: %d workers %d", inputSize, workers), func(t *testing.T) {
				got := s.request(input)
				assert(t, got, want)
			})
		}
	}
}

func helperParallel(ctx context.Context, s *service, input []int, resultC chan []int) func() error {
	return func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			r := s.request(input)
			fmt.Println(r)
			resultC <- r
			return nil
		}
	}
}

func TestBackgroundInParallel(t *testing.T) {
	input := buildInput(1000)
	want := buildOutput(input)

	numberOfRequestInParalell := 100
	numberOfWorkers := 16
	s := newService(numberOfWorkers)

	g, ctx := errgroup.WithContext(context.Background())
	resultC := make(chan []int, numberOfRequestInParalell)
	for i := 0; i < numberOfRequestInParalell; i++ {
		g.Go(helperParallel(ctx, s, input, resultC))
	}
	g.Wait()
	close(resultC)

	for got := range resultC {
		assert(t, got, want)
	}
}
