package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

func getHTTPRequest(ctx context.Context, value, sleepTime, group int) error {
	// simulate an expensive operation / call
	time.Sleep(time.Duration(rand.Int()%sleepTime) * time.Millisecond)

	// we can fail one group for test purposes
	if group == 1 && value == 4 {
		return fmt.Errorf("%d failed", group)
	}

	// i.e, i used a client to demonstrate the pattern on how to propagate the ctx of a process
	// the ctx should be propagated downstream, by convention first paramater of a function
	// if our ctx gets cancelled at any point, all the subrequest/subprocess will also be terminated
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://google.com", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func tryGetHTTPRequest(ctx context.Context, index, group int) func() error {
	return func() error {
		select {
		// propagates errors through the context
		case <-ctx.Done():
			fmt.Println(group, index, "cancelled")
			return ctx.Err()
		default:
			// main logic, note that if a context is cancelled it won't reach this path
			err := getHTTPRequest(ctx, index, 10000, group)
			if err != nil {
				return err
			}
			fmt.Println(group, index)
			return nil
		}
	}
}

// main
func multipleContext() {
	var cs int
	flag.IntVar(&cs, "c", 10, "number of corroutines per group")
	flag.Parse()

	// imagine we have two context and two groups
	g1, ctx1 := errgroup.WithContext(context.Background())
	g2, ctx2 := errgroup.WithContext(context.Background())

	for i := 0; i < cs; i++ {
		g1.Go(tryGetHTTPRequest(ctx1, i, 1))
	}
	for i := 0; i < cs; i++ {
		g2.Go(tryGetHTTPRequest(ctx2, i, 2))
	}

	if err := g1.Wait(); err != nil {
		fmt.Println(err)
	}
	if err := g2.Wait(); err != nil {
		fmt.Println(err)
	}

}
