package metrics_export

type WorkerMetrics struct {
	cpuUtil     float32
	memUsage    float32
	memCapacity float32
	appMetrics  map[string]AppMetrics
}

type AppMetrics struct {
	cpuUtil  float32
	memUsage float32
}
