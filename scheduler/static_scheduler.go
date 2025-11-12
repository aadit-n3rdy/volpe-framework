package scheduler

import (
	vcomms "volpe-framework/comms/volpe"

	"github.com/rs/zerolog/log"
)

type StaticScheduler struct {
	problems []string
	workers  []string
}


func NewStaticScheduler(problems []string) (*StaticScheduler, error) {
	sched := &StaticScheduler{}
	sched.problems = make([]string, len(problems))
	sched.workers = make([]string, 0)
	copy(sched.problems, problems)
	return sched, nil
}

func (ss *StaticScheduler) Init() error { return nil }

func (ss *StaticScheduler) AddWorker(worker string) {
	ss.workers = append(ss.workers, worker)
}

func (ss *StaticScheduler) UpdateMetrics(metrics *vcomms.MetricsMessage) {
	log.Warn().Caller().Msgf("skipping metrics update for static scheduler")
}

func (ss *StaticScheduler) RemoveWorker(worker string) {
	workerInd := -1
	for i, w := range ss.workers {
		if w == worker {
			workerInd = i
			break
		}
	}
	for i := workerInd; i < len(ss.workers); i++ {
		ss.workers[i] = ss.workers[i+1]
	}
	ss.workers = ss.workers[:len(ss.workers)-1]
}

func (ss *StaticScheduler) FillSchedule(sched Schedule) error {
	sched.Reset()
	for _, p := range ss.problems {
		for _, w := range ss.workers {
			sched.Set(w, p, 1) 
		}
	}
	return nil
}
