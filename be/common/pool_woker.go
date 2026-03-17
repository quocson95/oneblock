package common

import (
	"sync"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

var pool *ants.MultiPool

var poolOnce sync.Once

func GetWorkerPool() *ants.MultiPool {
	poolOnce.Do(func() {
		var err error
		pool, err = ants.NewMultiPool(10, 2000, ants.RoundRobin)
		if err != nil {
			panic(err)
		}
	})
	return pool
}

func AddTaskToWorker(task func()) {
	err := GetWorkerPool().Submit(task)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Worker pool is full")
	}
}
