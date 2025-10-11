package container_mgr

import (
	"context"
	"fmt"
	"math/rand/v2"
	comms "volpe-framework/comms/common"
	ccomms "volpe-framework/comms/container"
	vcomms "volpe-framework/comms/volpe"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProblemContainer struct {
	problemID     string
	containerName string
	containerPort uint16
	hostPort      uint16
	commsClient   ccomms.VolpeContainerClient
}

func genContainerName(problemID string) string {
	return fmt.Sprintf("volpe_%s_%d", problemID, rand.Int32())
}

const DEFAULT_CONTAINER_PORT uint16 = 8081

func NewProblemContainer(problemID string, imagePath string) (*ProblemContainer, error) {
	pc := new(ProblemContainer)
	pc.problemID = problemID
	pc.containerName = genContainerName(problemID)

	hostPort, err := runImage(imagePath, pc.containerName, DEFAULT_CONTAINER_PORT)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}
	pc.hostPort = hostPort

	cc, err := grpc.NewClient(fmt.Sprintf("localhost:%d", int(hostPort)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}
	pc.commsClient = ccomms.NewVolpeContainerClient(cc)

	return pc, nil
}

func (pc *ProblemContainer) GetContainerName() string {
	return pc.containerName
}

func (pc *ProblemContainer) GetSubpopulation() (*comms.Population, error) {
	pop, err := pc.commsClient.GetBestPopulation(context.Background(), &ccomms.PopulationSize{Size: -1})
	if err != nil {
		log.Err(err).Caller().Msg("")
		return nil, err
	}
	pop.ProblemID = &pc.problemID
	return pop, nil
}

func (pc *ProblemContainer) HandleEvents(eventChannel chan *vcomms.AdjustPopulationMessage) {
	for {
		msg, ok := <-eventChannel
		if !ok {
			log.Warn().Caller().Msgf("event channel for problemID %s was closed", pc.problemID)
			return
		}
		popSize := &ccomms.PopulationSize{Size: msg.GetSize()}
		reply, err := pc.commsClient.AdjustPopulationSize(context.Background(), popSize)
		if err != nil {
			log.Error().Caller().Msg(err.Error() + ", reply: " + reply.GetMessage())
			continue
		}

		pop := &comms.Population{Members: msg.GetSeed().GetMembers()}
		reply, err = pc.commsClient.InitFromSeedPopulation(context.Background(), pop)
		if err != nil {
			log.Error().Caller().Msg(err.Error() + ", reply: " + reply.GetMessage())
			continue
		}
	}
}
