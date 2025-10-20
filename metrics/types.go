package metrics

import (
	vcomm "volpe-framework/comms/volpe"
)

type WorkerMetrics struct {
	CpuUtil            float32
	MemUsage           float32
	MemTotal           float32
	ApplicationMetrics map[string]*ApplicationMetrics
}

func (wm *WorkerMetrics) ToProtobuf() vcomm.MetricsMessage {
	amMap := make(map[string]*vcomm.ApplicationMetrics)
	for key, val := range wm.ApplicationMetrics {
		amMap[key] = val.ToProtobuf()
	}
	return vcomm.MetricsMessage{
		CpuUtil:            wm.CpuUtil,
		MemUsage:           wm.MemUsage,
		MemTotal:           wm.MemTotal,
		ApplicationMetrics: amMap,
	}
}

type ApplicationMetrics struct {
	CpuUtil  float32
	MemUsage float32
}

func (am *ApplicationMetrics) ToProtobuf() *vcomm.ApplicationMetrics {
	res := vcomm.ApplicationMetrics{
		CpuUtil:  am.CpuUtil,
		MemUsage: am.MemUsage,
	}
	return &res
}
