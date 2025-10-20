package container_mgr

import (
	"context"
	"errors"
	"sync"
	ccoms "volpe-framework/comms/container"
	"volpe-framework/comms/volpe"

	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
)

// Manages an entire set of containers
// TODO: testing for this module

type ContainerManager struct {
	problemContainers map[string]*ProblemContainer
	pcMut             sync.Mutex
	containers        map[string]string
	containersUpdated bool
	meter             otelmetric.Meter
}

func NewContainerManager() *ContainerManager {
	cm := new(ContainerManager)
	cm.meter = otel.Meter("volpe-framework")
	return cm
}

func (cm *ContainerManager) AddProblem(problemID string, imagePath string) error {
	cm.pcMut.Lock()
	defer cm.pcMut.Unlock()
	_, ok := cm.problemContainers[problemID]
	if ok {
		log.Warn().Caller().Msgf("retried creating PC for pID %s, ignoring", problemID)
		// WARN: if supporting updating container, must change cm.containers here
		return errors.New("problemID already has container")
	}

	pc, err := NewProblemContainer(problemID, imagePath)
	if err != nil {
		log.Error().Caller().Msgf("error starting pID %s with image %s: %s", problemID, imagePath, err.Error())
		return err
	}
	cm.problemContainers[problemID] = pc
	cm.containers[pc.containerName] = problemID
	cm.containersUpdated = true
	return nil
}

func (cm *ContainerManager) HandlePopulationEvents(eventChannel chan *volpe.AdjustPopulationMessage) {
	for {
		msg, ok := <-eventChannel
		if !ok {
			log.Error().Caller().Msgf("event channel to CM closed")
			return
		}
		cm.handleEvent(msg)
	}
}

func (cm *ContainerManager) handleEvent(event *volpe.AdjustPopulationMessage) {
	cm.pcMut.Lock()
	defer cm.pcMut.Unlock()
	pc, ok := cm.problemContainers[event.GetProblemID()]
	if !ok {
		log.Warn().Caller().Msgf("received msg for problem ID %s, but problem container does not exist", event.GetProblemID())
		return
	}
	popSize := &ccoms.PopulationSize{Size: event.Size}
	_, err := pc.commsClient.AdjustPopulationSize(context.Background(), popSize)
	if err != nil {
		log.Error().Caller().Msgf("pop size adjust for pid %s got error %s", event.GetProblemID(), err.Error())
		return
	}
	_, err = pc.commsClient.InitFromSeedPopulation(context.Background(), event.GetSeed())
	if err != nil {
		log.Error().Caller().Msgf("pop seed for pid %s got error %s", event.GetProblemID(), err.Error())
		return
	}
}
