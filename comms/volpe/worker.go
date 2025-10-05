package volpe

import (
	"context"
	"volpe-framework/comms/common"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WorkerComms struct {
	client VolpeMasterClient
	stream grpc.BidiStreamingClient[WorkerMessage, MasterMessage]
	// TODO: include something for population
}

func NewWorkerComms(endpoint string, workerID string) (*WorkerComms, error) {
	// TODO: channel or something for population adjust
	wc := new(WorkerComms)
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}
	wc.client = NewVolpeMasterClient(conn)

	wc.stream, err = wc.client.StartStreams(context.Background())
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}

	err = wc.stream.Send(&WorkerMessage{
		Message: &WorkerMessage_WorkerID{
			WorkerID: &WorkerID{
				Id: workerID,
			},
		},
	})
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}

	return wc, nil
}

func (wc *WorkerComms) SendMetrics(metrics *MetricsMessage) error {
	workerMsg := WorkerMessage{Message: &WorkerMessage_Metrics{metrics}}
	err := wc.stream.Send(&workerMsg)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
	}
	return err
}

func (wc *WorkerComms) SendSubPopulation(population *common.Population) error {
	workerMsg := WorkerMessage{Message: &WorkerMessage_Population{population}}
	err := wc.stream.Send(&workerMsg)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
	}
	return err
}
