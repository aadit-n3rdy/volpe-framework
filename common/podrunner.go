package common

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/containers/common/libnetwork/types"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func RunPod(imagePath *string, containerName *string, containerPort uint16) (uint16, error) {
	conn, err := bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error creating podman connection: %s", err.Error()))
		return 0, err
	}
	f, err := os.Open(*imagePath)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error opening file %s: %s", *imagePath, err.Error()))
		return 0, err
	}
	imgReport, err := images.Import(conn, f, nil)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error importing image from %s: %s", *imagePath, err.Error()))
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("Loaded image %s from file %s\n", imgReport.Id, imagePath))
	sg := specgen.NewSpecGenerator(imgReport.Id, false)
	if containerName != nil {
		sg.Name = *containerName
	} else {
		sg.Name = "volpe_container"
	}
	hostPort, err := getFreePort()
	if err != nil {
		log.Error().Caller().Msg("error finding free host port: %s" + err.Error())
		return 0, err
	}
	pm := types.PortMapping{ContainerPort: containerPort, HostPort: uint16(hostPort)}
	sg.PortMappings = []types.PortMapping{pm}
	cr, err := containers.CreateWithSpec(conn, sg, nil)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("could not create container with image %s: %s", imgReport.Id, err.Error()))
		return 0, err
	}
	err = containers.Start(conn, cr.ID, nil)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("could not start container %s: %s", cr.ID, err.Error()))
		return 0, err
	}
	log.Info().Msg("successfully running container from " + *imagePath)
	return uint16(hostPort), nil
}
