package main

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

type service struct {
}

type fetchSingleDocumentJob struct {
	ctx            context.Context
	g              *errgroup.Group
	job            int
	resultsChannel chan<- int
	wg             *sync.WaitGroup
}

var jobs chan fetchSingleDocumentJob

func process(ctx context.Context, job int, resultsChannel chan<- int, done chan<- bool) func() error {
	return func() error {
		defer func() {
			done <- true
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resultsChannel <- job * 2
			return nil
		}
	}
}

func worker(jobs chan fetchSingleDocumentJob) func() error {
	done := make(chan bool)
	for job := range jobs {
		job.g.Go(process(job.ctx, job.job, job.resultsChannel, done))
		job.wg.Done()
		<-done
	}
	return nil
}

func newService(workers int) *service {
	s := &service{}
	jobs = make(chan fetchSingleDocumentJob)
	for i := 0; i < workers; i++ {
		go worker(jobs)
	}
	return s
}

func (s *service) request(input []int) []int {
	if len(input) == 0 {
		return []int{}
	}

	results := make(chan int, len(input))

	g, ctx := errgroup.WithContext(context.Background())
	var wg sync.WaitGroup

	wg.Add(len(input))
	for _, v := range input {
		go func(v int) {
			fetchSingleDocumentJob := fetchSingleDocumentJob{
				g:              g,
				ctx:            ctx,
				job:            v,
				resultsChannel: results,
				wg:             &wg,
			}
			jobs <- fetchSingleDocumentJob
		}(v)
	}
	wg.Wait()

	// Do we need to wait for the groups of jobs to finish?
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
	close(results)

	return buildResult(results)
}
