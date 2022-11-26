package alias

import (
	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/alias"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"
	manager "github.com/kernelschmelze/pkg/plugin/manager"
)

type config struct {
}

type Plugin struct {
	*plugin.PluginBase
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(899),
	}

	err := p.Init(plugin.PluginConfig{
		p,
		p.start,
		p.stop,
		p.configure,
		p.do,
		&config{},
	})

	if err != nil {
		log.Errorf("init plugin '%T' failed, err=%s", p, err)
	}

	return p
}

func (p *Plugin) start() error {

	log.Infof("start %T", p)

	return nil
}

func (p *Plugin) stop() error {

	log.Infof("stop %T", p)

	return nil
}

func (p *Plugin) configure(v interface{}) {

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		message := alias.Get(data)

		manager.Dispatch(message)

	}

	return nil
}
