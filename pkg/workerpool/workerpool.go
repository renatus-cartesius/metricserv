package workerpool

import "sync"

type Pool struct {
	wg *sync.WaitGroup
}

func NewPool() (*Pool, error) {
	return &Pool{
		wg: &sync.WaitGroup{},
	}, nil
}

func (p *Pool) Listen(workers int, do func()) {

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			do()
		}()
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}
