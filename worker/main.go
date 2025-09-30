package main

import (
	"context"
	"time"

	cm "volpe-framework/container_mgr"
	gc "volpe-framework/grpc_comms"
	met "volpe-framework/metrics_export"

	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	met.InitOTelSDK()
	containerName := "grpc_comms_test"
	problems := map[string]string{
		"grpc_comms": "grpc_comms_test",
	}
	port, err := cm.RunImage("../grpc_comms/grpc_test_img.tar", containerName, 8081)
	if err != nil {
		log.Fatal().Caller().Msgf("failed to run pod with error: %s", err.Error())
		panic(err)
	}
	defer cm.StopContainer(containerName)
	log.Log().Caller().Msgf("started container at port %d", port)

	pme, err := met.NewPodmanMetricExporter("test_device", problems)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
	}
	go pme.Run()
	log.Log().Caller().Msgf("started OTel metric exporter")

	cc, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Caller().Msgf("failed to create grpc client: %s", err.Error())
		panic(err)
	}
	client := gc.NewVolpeContainerClient(cc)
	for {
		resp, err := client.SayHello(context.Background(), &gc.HelloRequest{Name: "xyz"})
		if err != nil {
			log.Fatal().Caller().Msgf("failed to call sayhello: %s", err.Error())
			panic(err)
		}
		if resp.GetMessage() != "hello xyz" {
			log.Fatal().Caller().Msgf("not expected msg: %s", resp.GetMessage())
			panic(err)
		} else {
			log.Log().Caller().Msg("got expected msg")
		}
		time.Sleep(5 * time.Second)
	}

}
