package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

var meterprovider *metric.MeterProvider
var exporter *otlpmetricgrpc.Exporter
var reader *metric.PeriodicReader

func InitOTelSDK() error {
	var err error
	exporter, err = otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return err
	}
	reader = metric.NewPeriodicReader(exporter, metric.WithInterval(3*time.Second))
	meterprovider = metric.NewMeterProvider(metric.WithReader(reader))

	otel.SetMeterProvider(meterprovider)
	return nil
}

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
		CpuUtil:            0,
		MemUsage:           0,
		MemTotal:           0.01,
		ApplicationMetrics: make(map[string]*ApplicationMetrics),
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
	worker.metrics = WorkerMetrics{
		CpuUtil:            args.metrics.CpuUtil,
		MemUsage:           args.metrics.MemUsage,
		MemTotal:           args.metrics.MemTotal,
		ApplicationMetrics: args.metrics.ApplicationMetrics,
	}
	worker.lastUpdate = time.Now()
	return nil
}
