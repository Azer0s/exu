package exu

import "sync"

var globalWaitGroup = sync.WaitGroup{}

func WithWaitGroup(f func()) {
	globalWaitGroup.Add(1)

	go func() {
		f()
		globalWaitGroup.Done()
	}()
}

func AllSettled() {
	globalWaitGroup.Wait()
}
