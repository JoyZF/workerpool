package main

import (
	"github.com/JoyZF/workerpool"
	"time"

)

func main()  {
	p := workerpool.New(5, workerpool.WithPreAllocWorker(false), workerpool.WithBlock(false))

	for i := 0; i < 10; i++ {
		err := p.Schedule(func() {
			time.Sleep(time.Second * 3)
		})
		if err != nil {
			println("task: ", i, "err:", err)
		}
	}
	p.Free()
}