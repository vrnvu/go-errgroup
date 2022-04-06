package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

func print(ctx context.Context, i int) func() error {
	return func() error {
		r := (rand.Int() % 100)
		time.Sleep(time.Duration(r) * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Printf("number = %d\n", i)
			if i == 4 {
				return errors.New("dep")
			}
			return nil
		}
	}
}

func main() {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < 10; i++ {
		g.Go(print(ctx, i))
	}

	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
}
