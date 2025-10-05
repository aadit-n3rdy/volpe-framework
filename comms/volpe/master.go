package volpe

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type MasterComms struct {
	mcs masterCommsServer
	sr  *grpc.Server
	lis net.Listener
}

type masterCommsServer struct {
	chans_mut  sync.RWMutex
	channs     map[string]chan *MasterMessage
	metricChan chan *MetricsMessage
	// TODO: include something for population
}

func mcsStreamHandlerThread(
	workerID string,
	stream grpc.BidiStreamingServer[WorkerMessage, MasterMessage],
	masterSendChan chan *MasterMessage,
	metricChan chan *MetricsMessage,
	// TODO: add a population handler here, to send popln msgs
) {

	masterRecvChan := make(chan *WorkerMessage)
	var closeReader sync.Mutex
	closeReader.Lock()

	readerThread := func() {
		for {
			if closeReader.TryLock() {
				close(masterRecvChan)
				return
			}
			wm, err := stream.Recv()
			if err != nil {
				log.Error().Caller().Msg(err.Error())
				close(masterRecvChan)
				return
			}
			masterRecvChan <- wm
		}
	}
	go readerThread()
	for {
		select {
		case result, ok := <-masterRecvChan:
			if !ok {
				// TODO: Notify of stream closure
				return
			}
			if result.GetMetrics() != nil {
				metricChan <- result.GetMetrics()
			} else if result.GetPopulation() != nil {
				// TODO: Send population to appropriate worker
			} else if result.GetWorkerID() != nil {
				log.Warn().Caller().Msg("got workerID from stream for " + workerID)
			}
		case result, ok := <-masterSendChan:
			if !ok {
				closeReader.Unlock()
			}
			err := stream.Send(result)
			if err != nil {
				log.Error().Caller().Msg(err.Error())
				// readerThread will close automatically, JIC
				closeReader.Unlock()
				// TODO: inform that stream no longer works
				return
			}
		}
	}
}

func initMasterCommsServer(mcs *masterCommsServer, metricChan chan *MetricsMessage /* TODO: popln */) (err error) {
	mcs.channs = make(map[string]chan *MasterMessage)
	mcs.metricChan = metricChan
	return nil
}

func (mcs *masterCommsServer) StartStreams(stream grpc.BidiStreamingServer[WorkerMessage, MasterMessage]) error {
	protoMsg, err := stream.Recv()
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return err
	}
	workerIdMsg := protoMsg.GetWorkerID()
	if workerIdMsg == nil {
		return errors.New("expected WorkerID msg first")
	}
	workerID := workerIdMsg.GetId()

	masterSendChan := make(chan *MasterMessage)

	mcs.chans_mut.Lock()
	defer mcs.chans_mut.Unlock()

	mcs.channs[workerID] = masterSendChan

	go mcsStreamHandlerThread(workerID, stream, masterSendChan, mcs.metricChan)

	return nil
}

func (mcs *masterCommsServer) mustEmbedUnimplementedVolpeMasterServer() {}

func NewMasterComms(port uint16, metricChan chan *MetricsMessage /* TODO: include something for popln */) (*MasterComms, error) {
	mc := new(MasterComms)
	err := initMasterCommsServer(&mc.mcs, metricChan)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}
	sr := grpc.NewServer()
	mc.sr = sr
	RegisterVolpeMasterServer(sr, &mc.mcs)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}
	mc.lis = lis

	return mc, nil
}

func (mc *MasterComms) Serve() error {
	err := mc.sr.Serve(mc.lis)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
	}
	return err
}
