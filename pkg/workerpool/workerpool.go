package workerpool

import (
	"context"
	"runtime"
	"sync"

	"github.com/google/uuid"
	"github.com/renatus-cartesius/metricserv/internal/logger"
	"go.uber.org/zap"
)

type Pool[T any] struct {
	wg    *sync.WaitGroup
	jobCh chan T
}

func NewPool[T any]() (*Pool[T], error) {
	return &Pool[T]{
		wg:    &sync.WaitGroup{},
		jobCh: make(chan T, runtime.NumCPU()),
	}, nil
}

func (p *Pool[T]) Listen(ctx context.Context, workers int, do func(T)) {

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go func(ctx context.Context) {
			defer p.wg.Done()
			id := uuid.NewString()
			logger.Log.Debug(
				"creating worker",
				zap.String("uuid", id),
			)
			for {
				select {
				case job, ok := <-p.jobCh:
					if !ok {
						logger.Log.Debug(
							"closed worker",
							zap.String("uuid", id),
						)
						return
					}
					do(job)
				case <-ctx.Done():
					logger.Log.Debug(
						"closed worker",
						zap.String("uuid", id),
					)
					return
				}
			}
		}(ctx)
	}
}

func (p *Pool[T]) Wait() {
	p.wg.Wait()
}

func (p *Pool[T]) AddJob(job T) {
	p.jobCh <- job
}

func (p *Pool[T]) Stop() {
	close(p.jobCh)
}
