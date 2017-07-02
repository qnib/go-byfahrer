package byfahrer

import (
	"fmt"
	"reflect"

	"github.com/zpatrick/go-config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/qnib/qframe-types"
	"strings"
)

const (
	version   = "0.0.0"
	pluginTyp = qtypes.HANDLER
	pluginPkg = "go-byfahrer"
	dockerAPI = "v1.29"

)

type Plugin struct {
	qtypes.Plugin
	engCli *client.Client
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	var err error
	p := Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}
	return p, err
}

func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start plugin v%s", p.Version))
	dockerHost := p.CfgStringOr("docker-host", "unix:///var/run/docker.sock")
	// Filter start/stop event of a container
	var err error
	p.engCli, err = client.NewClient(dockerHost, dockerAPI, nil, nil)
	if err != nil {
		p.Log("error", fmt.Sprintf("Could not connect docker/docker/client to '%s': %v", dockerHost, err))
		return
	}
	info, err := p.engCli.Info(context.Background())
	if err != nil {
		p.Log("error", fmt.Sprintf("Error during Info(): %v >err> %s", info, err))
		return
	} else {
		p.Log("info", fmt.Sprintf("Connected to '%s' / v'%s'", info.Name, info.ServerVersion))
	}
	bg := p.QChan.Data.Join()
	for {
		val := bg.Recv()
		switch val.(type) {
		case qtypes.ContainerEvent:
			ce := val.(qtypes.ContainerEvent)
			if ce.Event.Type != "container" {
				continue
			}
			if strings.HasPrefix(ce.Event.Action, "health_status") {
				continue
			}
			switch ce.Event.Action {
			case "exec_create", "exec_start", "resize":
				continue
			case "start":
				go p.createProxy(ce.Container)
			case "die":
			default:
				p.Log("info", fmt.Sprintf("Go '%s': %s.%s", ce.Event.Actor.ID, ce.Event.Type, ce.Event.Action))
			}
		default:
			p.Log("warn", fmt.Sprintf("Dunno message type: %s", reflect.TypeOf(val)))
		}
	}
}

func (p *Plugin) createProxy(cnt types.ContainerJSON) {
	imageName, ok := cnt.Config.Labels["org.qnib.go-byfahrer.proxy-image"]
	if !ok {
		p.Log("debug", fmt.Sprintf("Undefined label 'org.qnib.byfahrer.proxy-image' %v", cnt.Config.Labels))
		return
	} else {
		p.Log("info", fmt.Sprintf("Use org.qnib.byfahrer.proxy-image=%s to start proxy", imageName))
	}
	cntCfg := &container.Config{
		AttachStdin: false,
		AttachStdout: false,
		AttachStderr: false,
		Image: imageName,
	}
	nwMode := fmt.Sprintf("container:%s", cnt.ID)
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(nwMode),
	}
	networkingConfig := &network.NetworkingConfig{}
	cntName := strings.TrimLeft(cnt.Name, "/")
	container, err := p.engCli.ContainerCreate(context.Background(), cntCfg, hostConfig, networkingConfig, fmt.Sprintf("%s-proxy", cntName))
	if err != nil {
		p.Log("error", fmt.Sprintf("Failed to create container: %s", err.Error()))
		return
	}
	p.Log("info", fmt.Sprintf("Create proxy container '%s-proxy' for '%s'", cntName, cntName))
	err = p.engCli.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		p.Log("error", fmt.Sprintf("Failed to start container: %s", err.Error()))
		return
	}

}
