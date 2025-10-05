package volpe_test

import (
	"testing"

	comms "volpe-framework/comms/volpe"
)

// Tests comms between master and worker

func TestComms(t *testing.T) {
	// 1. setup master

	metricChan := make(chan *comms.MetricsMessage, 5)

	mc, err := comms.NewMasterComms(8118, metricChan)
	if err != nil {
		t.Fatal(err)
	}
	go mc.Serve()

	// 2. setup worker
	wc, err := comms.NewWorkerComms("localhost:8118", "abcd123")
	if err != nil {
		t.Fatal(err)
	}

	// 3. send metrics on worker, check if received by master
	metricsMsg := comms.MetricsMessage{
		CpuUtil:  1.5,
		MemUsage: 1.5,
		MemTotal: 1.5,
		ApplicationMetrics: map[string]*comms.ApplicationMetrics{
			"application1": {
				CpuUtil:  1.5,
				MemUsage: 1.0,
			},
		},
		WorkerID: "abcd123",
	}
	err = wc.SendMetrics(&metricsMsg)
	if err != nil {
		t.Fatal(err)
	}

	recMsg, ok := <-metricChan
	if !ok {
		t.Fatal("channel closed")
	}
	if recMsg.CpuUtil != 1.5 || recMsg.MemUsage != 1.5 || recMsg.WorkerID != "abcd123" {
		t.Fail()
	}
}
