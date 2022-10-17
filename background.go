package main

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

type service struct {
}

type jobRequest struct {
	ctx      context.Context
	g        *errgroup.Group
	job      int
	resultsC chan<- int
	wg       *sync.WaitGroup
}

var jobsC chan jobRequest

func process(ctx context.Context, job int, resultsC chan<- int, done chan<- bool) func() error {
	return func() error {
		defer func() {
			done <- true
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resultsC <- job * 2
			return nil
		}
	}
}

func worker() func() error {
	done := make(chan bool)
	for job := range jobsC {
		job.g.Go(process(job.ctx, job.job, job.resultsC, done))
		job.wg.Done()
		<-done
	}
	return nil
}

func newService(workers int) *service {
	s := &service{}
	jobsC = make(chan jobRequest)
	for i := 0; i < workers; i++ {
		go worker()
	}
	return s
}

func (s *service) request(input []int) []int {
	if len(input) == 0 {
		return []int{}
	}

	resultsC := make(chan int, len(input))

	g, ctx := errgroup.WithContext(context.Background())
	var wg sync.WaitGroup

	// we publish jobs to the channel
	wg.Add(len(input))
	for _, v := range input {
		go func(v int) {
			jobRequest := jobRequest{
				g:        g,
				ctx:      ctx,
				job:      v,
				resultsC: resultsC,
				wg:       &wg,
			}
			jobsC <- jobRequest
		}(v)
	}
	// we wait for all the jobs to start
	wg.Wait()

	// we wait for all the jobs to finish
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
	close(resultsC)

	return buildResult(resultsC)
}
