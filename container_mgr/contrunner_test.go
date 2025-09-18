package container_mgr_test

import (
	"context"
	"testing"

	cm "volpe-framework/container_mgr"
	gc "volpe-framework/grpc_comms"

	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRunImage(t *testing.T) {
	containerName := "grpc_comms_test"
	port, err := cm.RunImage("../grpc_comms/grpc_test_img.tar", containerName, 8081)
	if err != nil {
		t.Fatalf("failed to run pod with error: %s", err.Error())
	}
	defer cm.StopContainer(containerName)
	t.Logf("started container at port %d", port)
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
}
