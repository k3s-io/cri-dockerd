package container

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type CreateConfigWrapper struct {
	Config           *container.Config
	Name             string
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}
