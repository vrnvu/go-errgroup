package main

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/sync/errgroup"
)

func processJob(ctx context.Context, workerID int, job int, results chan<- int, done chan bool) func() error {
	return func() error {
		// if error we need to manually call the done channel!
		// so we ALWAYS need to call this
		// a defer func() is probably the easiest
		defer func() {
			done <- true
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// if job == 2 {
			// 	return errors.New("job was 2")
			// }
			results <- job * 2
			return nil
		}
	}

}

func coordinatedWorker(ctx context.Context, wctx context.Context, workerID int, g *errgroup.Group, jobs <-chan int, results chan<- int) func() error {
	return func() error {
		select {
		case <-wctx.Done():
			return wctx.Err()
		default:
			done := make(chan bool)
			for job := range jobs {
				g.Go(processJob(ctx, workerID, job, results, done))
				<-done
			}
			return nil
		}
	}
}

func buildResult(results <-chan int) []int {
	r := make([]int, len(results))
	i := 0
	for result := range results {
		r[i] = result
		i++
	}
	sort.Ints(r)
	return r
}

// Workers run an example of corroutine coordination
// We initialize the number of workers specified in the argument
// For [1,2,3] we will return [2,4,6]
func Workers(workers int, input []int) []int {
	numJobs := len(input)

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	wg, wctx := errgroup.WithContext(context.Background())
	g, ctx := errgroup.WithContext(context.Background())

	// start workers
	for i := 0; i < workers; i++ {
		wg.Go(coordinatedWorker(ctx, wctx, i, g, jobs, results))
	}

	// send jobs
	for _, v := range input {
		jobs <- v
	}
	close(jobs)

	// we need to wait for the worker group FIRST in this example
	if err := wg.Wait(); err != nil {
		fmt.Println(err)
	}

	// Do we need to wait for the groups of jobs to finish?
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
	close(results)

	return buildResult(results)
}
