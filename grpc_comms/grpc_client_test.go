package grpc_comms_test

import (
	"context"
	"testing"

	gc "volpe-framework/grpc_comms"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGrpcComms(t *testing.T) {
	t.Log("Using port 8081 for testing")
	cc, err := grpc.NewClient("localhost:35209", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("error connecting to localhost:8081, %s", err.Error())
	}
	conn := gc.NewVolpeContainerClient(cc)
	res, err := conn.SayHello(context.Background(), &gc.HelloRequest{Name: "xyz"})
	if err != nil {
		t.Fatalf("failed in calling grpc, %s", err.Error())
	}
	t.Logf("received msg %s", res.GetMessage())
	if res.GetMessage() != "hello xyz" {
		t.Errorf("received msg %s was not expected msg", res.GetMessage())
	}
}
