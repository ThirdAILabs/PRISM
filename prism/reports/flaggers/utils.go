package flaggers

import "sync"

type CompletedTask[T any] struct {
	Result T
	Err    error
}

func RunInPool[In any, Out any](worker func(In) (Out, error), queue chan In, maxWorkers int) chan CompletedTask[Out] {
	workers := min(len(queue), maxWorkers)

	completed := make(chan CompletedTask[Out], len(queue))

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				
				for {
					next, ok := <-queue
					if !ok {
						return
					}

					res, err := worker(next)
					if err != nil {
						completed <- CompletedTask[Out]{Err: err}
					} else {
						completed <- CompletedTask[Out]{Result: res, Err: nil}
					}
				}
			}()
		}

		wg.Done()

		close(completed)
	}()

	return completed
}
