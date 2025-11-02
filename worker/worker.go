package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	vcomms "volpe-framework/comms/volpe"
	contman "volpe-framework/container_mgr"
	"volpe-framework/metrics"

	"github.com/rs/zerolog/log"
)

func main() {
	metrics.InitOTelSDK()
	endpoint, ok := os.LookupEnv("VOLPE_MASTER")
	if !ok {
		log.Warn().Caller().Msgf("using default VOLPE_MASTER of localhost:8080")
		endpoint = "localhost:8080"
	}
	workerID, ok := os.LookupEnv("VOLPE_WORKER_ID")
	if !ok {
		workerID = "worker_" + fmt.Sprintf("%d", rand.Int())
		log.Warn().Caller().Msgf("VOLPE_WORKER_ID not found, using %s instead", workerID)
	}

	wc, err := vcomms.NewWorkerComms(endpoint, workerID)
	if err != nil {
		log.Fatal().Caller().Msgf("could not create workercomms: %s", err.Error())
		panic(err)
	}
	cm := contman.NewContainerManager()

	go cm.RunMetricsExport(wc, workerID)

	err = cm.AddProblem("problem1", "../comms/pybindings/grpc_test_img.tar")
	if err != nil {
		log.Fatal().Caller().Msgf("failed to run pod with error: %s", err.Error())
		panic(err)
	}
	// TODO: stop container
	// defer cm.StopContainer(containerName)
	log.Log().Caller().Msgf("started container at port unknown") // %d", -1)

	for {
		time.Sleep(1000 * time.Millisecond)
	}

	// pme, err := met.NewPodmanMetricExporter("test_device", problems)
	// if err != nil {
	// 	log.Error().Caller().Msg(err.Error())
	// }
	// go pme.Run()
	// log.Log().Caller().Msgf("started OTel metric exporter")

	// cc, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port),
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatal().Caller().Msgf("failed to create grpc client: %s", err.Error())
	// 	panic(err)
	// }
	// client := gc.NewVolpeContainerClient(cc)
	// for {
	// 	resp, err := client.SayHello(context.Background(), &gc.HelloRequest{Name: "xyz"})
	// 	if err != nil {
	// 		log.Fatal().Caller().Msgf("failed to call sayhello: %s", err.Error())
	// 		panic(errors.New("failed to call sayhello"))
	// 	}
	// 	if resp.GetMessage() != "hello xyz" {
	// 		log.Fatal().Caller().Msgf("not expected msg: %s", resp.GetMessage())
	// 		panic(errors.New("unexpected msg from container"))
	// 	} else {
	// 		log.Log().Caller().Msg("got expected msg")
	// 	}
	// 	time.Sleep(5 * time.Second)
	// }

}
