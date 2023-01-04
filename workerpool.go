package workerpool

import (
	"errors"
	"fmt"
	"sync"
)

const (
	defaultCapacity = 1 << 8
	maxCapacity     = 1 << 20
)

var ErrWorkerPoolFreed = errors.New("workerpool freed")

type Pool struct {
	capacity int            // pool大小
	active   chan struct{}  // active channel
	tasks    chan Task      // task channel
	wg       sync.WaitGroup // 用于在pool销毁时等待所有worker退出
	quit     chan struct{}  // 用于通知各个worker退出信号channel
	preAlloc bool           // 是否在创建pool的时候就预创建workers，默认值为：false
	block    bool           // 当pool满的情况下，新的Schedule调用是否阻塞当前goroutine。默认值：true 如果block = false，则Schedule返回ErrNoWorkerAvailInPool
}

type Task func()

func New(capacity int, opts ...Option) *Pool {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}
	p := &Pool{
		capacity: capacity,
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
		active:   make(chan struct{}, capacity),
	}

	for _,opt := range opts {
		opt(p)
	}

	fmt.Println("[workerpool start]\n")

	if p.preAlloc {
		// create all goroutines and send into works channel
		for i := 0; i < p.capacity; i++ {
			p.newWorker(i + 1)
			p.active <- struct{}{}
		}
	}

	go p.run()
	return p
}

func (p *Pool) run() {
	idx := len(p.active)

	if !p.preAlloc {
		loop:
			for t := range p.tasks {
				p.returnTask(t)
				select {
				case <-p.quit:
					return
				case p.active <- struct{}{}:
					idx++
					p.newWorker(idx)
				default:
					break loop
				}
			}
	}


	for {
		select {
		case <-p.quit:
			fmt.Println("[workerpool quit]\n")
			return
		case p.active <- struct{}{}:
			idx++
			p.newWorker(idx)
		}
	}
}

func (p *Pool) newWorker(idx int) {
	p.wg.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("worker[%03d]: recover panic[%s] and exit\n", idx, err)
				<-p.active
			}
			p.wg.Done()
		}()
	}()

	fmt.Printf("worker[%03d]:start\n", idx)

	for {
		select {
		case <-p.quit:
			fmt.Printf("worker[%03d]: exit\n", idx)
			<-p.active
			return
		case t := <-p.tasks:
			fmt.Printf("worker[%03d]: receive a task\n", idx)
			idx++
			t()
		}
	}
}

func (p *Pool) Schedule(t Task) error {
	select {
	case <-p.quit:
		return ErrWorkerPoolFreed
	case p.tasks <- t:
		return nil
	}
}

func (p *Pool) Free() {
	close(p.quit)

	p.wg.Wait()
	fmt.Printf("workerpool freed\n")
}

func (p *Pool) returnTask(t Task) {
	t()
}
