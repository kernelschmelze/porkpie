package filter

import (
	"strconv"
	"strings"
	"sync"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"
)

type config struct {
	Pattern []string
}

type Plugin struct {
	*plugin.PluginBase
	pattern []string
	guard   sync.RWMutex
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(-1),
		[]string{},
		sync.RWMutex{},
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

	if config, ok := v.(*config); ok {

		p.guard.Lock()

		for i := range config.Pattern {

			pattern := config.Pattern[i]

			keys := strings.Split(pattern, ":")
			if len(keys) == 2 {
				p.pattern = append(p.pattern, pattern)
			}

		}

		p.guard.Unlock()

	}

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		p.guard.RLock()

		for i := range p.pattern {

			pattern := p.pattern[i]

			keys := strings.Split(pattern, ":")
			if len(keys) == 2 {

				if data.GetGID() == toUint64(keys[0]) && data.GetSID() == toUint64(keys[1]) {
					data.Drop = true
					break
				}
			}
		}

		p.guard.RUnlock()

	}

	return nil
}

func toUint64(data string) uint64 {

	data = strings.TrimSpace(data)
	if u64, err := strconv.ParseUint(data, 10, 32); err == nil {
		return u64
	}

	return 0
}
