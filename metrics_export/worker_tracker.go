package metrics_export

import (
	"sync"
	"time"
)

type StaticWorkerTracker struct {
	workers     map[WorkerID]*WorkerData
	workerMutex sync.RWMutex
}

type WorkerID string

type WorkerData struct {
	id         WorkerID
	metrics    WorkerMetrics
	lastUpdate time.Time
}

func newZeroMetrics() WorkerMetrics {
	return WorkerMetrics{
		cpuUtil:     0,
		memUsage:    0,
		memCapacity: 0.01,
		appMetrics:  make(map[string]AppMetrics),
	}
}

func NewStaticWorkerTracker(workerIDs []WorkerID) (swt *StaticWorkerTracker, err error) {
	err = nil
	swt = new(StaticWorkerTracker)
	swt.workers = make(map[WorkerID]*WorkerData)
	for i := range workerIDs {
		wd := WorkerData{
			id:         workerIDs[i],
			metrics:    newZeroMetrics(),
			lastUpdate: time.Unix(0, 0),
		}
		swt.workers[workerIDs[i]] = &wd
	}
	return
}

type SetMetricsRequest struct {
	id      WorkerID
	metrics WorkerMetrics
}

func (swt *StaticWorkerTracker) SetMetrics(args *SetMetricsRequest, res *int32) error {
	swt.workerMutex.Lock()
	defer swt.workerMutex.Unlock()
	worker := swt.workers[args.id]
	worker.metrics = args.metrics
	worker.lastUpdate = time.Now()
	return nil
}
