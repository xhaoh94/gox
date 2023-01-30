package discovery

import (
	"context"

	"github.com/xhaoh94/gox/types"
)

type (
	Discovery struct {
		actor   *ActorDiscovery
		service *ServiceDiscovery
	}
)

func New(engine types.IEngine, ctx context.Context) *Discovery {
	discovery := new(Discovery)
	discovery.actor = newActorDiscovery(engine, ctx)
	discovery.service = newServiceDiscovery(engine, ctx)
	return discovery
}

func (nw *Discovery) Service() types.IServiceDiscovery {
	return nw.service
}
func (nw *Discovery) Actor() types.IActorDiscovery {
	return nw.actor
}

func (nw *Discovery) Init() {
	nw.service.Start()
	nw.actor.Start()
}
func (nw *Discovery) Destroy() {
	nw.actor.Stop()
	nw.service.Stop()
}
