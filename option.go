package workerpool

type Option func(pool *Pool)

func WithBlock(block bool) Option {
	return func(pool *Pool) {
		pool.block = block
	}
}

func WithPreAllocWorker(preAlloc bool) Option {
	return func(pool *Pool) {
		pool.preAlloc = preAlloc
	}
}