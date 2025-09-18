package container_mgr

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

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

func RunImage(imagePath string, containerName string, containerPort uint16) (uint16, error) {
	uid := os.Getuid()
	conn, err := bindings.NewConnection(context.Background(), fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", uid))
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error creating podman connection: %s", err.Error()))
		return 0, err
	}

	dir, _ := os.Getwd()
	log.Warn().Msgf("running from %s", dir)

	f, err := os.Open(imagePath)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error opening img file: %s", err.Error()))
		return 0, err
	}

	imgLoadReport, err := images.Load(conn, f) //, &images.LoadOptions{Reference: &imageName})
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error importing image from %s: %s", imagePath, err.Error()))
		return 0, err
	}
	log.Warn().Msg(fmt.Sprintf("loaded image %s from file %s\n", imgLoadReport.Names[0], imagePath))

	imageName := imgLoadReport.Names[0]

	ignore := true
	_ = containers.Stop(conn, containerName, &containers.StopOptions{Ignore: &ignore})
	_, _ = containers.Remove(conn, containerName, &containers.RemoveOptions{Ignore: &ignore})

	sg := specgen.NewSpecGenerator(imageName, false)
	remove := true
	sg.Remove = &remove
	sg.Name = containerName
	pm := types.PortMapping{ContainerPort: containerPort}
	sg.PortMappings = []types.PortMapping{pm}
	cr, err := containers.CreateWithSpec(conn, sg, nil)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("could not create container with image %s: %s", imageName, err.Error()))
		return 0, err
	}
	err = containers.Start(conn, cr.ID, nil)
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("could not start container %s: %s", cr.ID, err.Error()))
		return 0, err
	}
	inspectContainerData, _ := containers.Inspect(conn, containerName, nil)
	hostPort, _ := strconv.Atoi(inspectContainerData.HostConfig.PortBindings[fmt.Sprintf("%d/tcp", containerPort)][0].HostPort)
	log.Info().Msg("successfully running container from " + imagePath)
	return uint16(hostPort), nil
}

func StopContainer(containerName string) {
	uid := os.Getuid()
	conn, err := bindings.NewConnection(context.Background(), fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", uid))
	if err != nil {
		log.Error().Caller().Msg(fmt.Sprintf("error creating podman connection: %s", err.Error()))
	}
	_ = containers.Stop(conn, containerName, nil)
}
