package main

import (
	"fmt"
	"os"
	ccomms "volpe-framework/comms/common"
	vcomms "volpe-framework/comms/volpe"
	cm "volpe-framework/container_mgr"
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

	cman := cm.NewContainerManager()

	metricChan := make(chan *vcomms.MetricsMessage, 10)
	popChan := make(chan *ccomms.Population, 10)

	mc, err := vcomms.NewMasterComms(portD, metricChan, popChan)
	if err != nil {
		log.Fatal().Caller().Msgf("error initializing master comms: %s", err.Error())
		panic(err)
	}

	go clearMetricChan(metricChan)

	go sendPopulation(cman, popChan)

	mc.Serve()
}

func sendPopulation(cman *cm.ContainerManager, popChan chan *ccomms.Population) {
	for {
		m, ok := <-popChan
		if !ok {
			log.Error().Caller().Msg("popChan closed")
			return
		}
		cman.IncorporatePopulation(m)
		log.Info().Caller().Msgf("received population for problem %s", m.GetProblemID())
	}
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
