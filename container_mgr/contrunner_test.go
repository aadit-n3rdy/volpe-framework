package container_mgr_test

import (
	"context"
	"testing"
	"time"

	gc "volpe-framework/comms/container"
	cm "volpe-framework/container_mgr"
	met "volpe-framework/metrics_export"

	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRunImage(t *testing.T) {
	containerName := "grpc_comms_test"
	problems := map[string]string{
		"grpc_comms": "grpc_comms_test",
	}
	port, err := cm.RunImage("../comms/pybindings/grpc_test_img.tar", containerName, 8081)
	if err != nil {
		t.Fatalf("failed to run pod with error: %s", err.Error())
	}
	defer cm.StopContainer(containerName)
	t.Logf("started container at port %d", port)

	pme, err := met.NewPodmanMetricExporter("test_device", problems)
	go pme.Run()
	t.Logf("started OTel metric exporter")

	cc, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create grpc client: %s", err.Error())
	}
	client := gc.NewVolpeContainerClient(cc)
	resp, err := client.SayHello(context.Background(), &gc.HelloRequest{Name: "xyz"})
	if err != nil {
		t.Fatalf("failed to call sayhello: %s", err.Error())
	}
	if resp.GetMessage() != "hello xyz" {
		t.Fatalf("not expected msg: %s", resp.GetMessage())
	} else {
		t.Log("got expected msg")
	}

	deadline, _ := t.Deadline()
	time.Sleep(time.Now().Sub(deadline.Add(-time.Second)))
}
