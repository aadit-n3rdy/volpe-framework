package main

import (
	"fmt"
	"os"
	vcomms "volpe-framework/comms/volpe"
	"volpe-framework/metrics"

	"github.com/rs/zerolog/log"
)

func main() {
	metrics.InitOTelSDK()

	port, ok := os.LookupEnv("VOLPE_PORT")
	if !ok {
		log.Warn().Caller().Msgf("using default VOLPE_PORT of 8080")
		port = "8080"
	}
	portD := uint16(0)
	fmt.Sscan(port, &portD)

	metricChan := make(chan *vcomms.MetricsMessage, 10)

	mc, err := vcomms.NewMasterComms(portD, metricChan)
	if err != nil {
		log.Fatal().Caller().Msgf("error initializing master comms: %s", err.Error())
		panic(err)
	}

	go clearMetricChan(metricChan)

	mc.Serve()
}

func clearMetricChan(metricChan chan *vcomms.MetricsMessage) {
	// dummy func to just read metrics
	for {
		m, ok := <-metricChan
		if !ok {
			log.Error().Caller().Msg("metricChan closed")
			return
		}
		log.Info().Caller().Msgf("cleared metric from %s", m.GetWorkerID())
	}
}
