package ip2asn

import (
	"sync"
	"time"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/pkg/plugin/watcher"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"
)

type config struct {
	DB string
}

type Plugin struct {
	*plugin.PluginBase
	guard  sync.RWMutex
	lookup map[string][]asn
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(501),
		sync.RWMutex{},
		make(map[string][]asn),
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

		if len(config.DB) > 0 {
			p.openDB(config.DB)
		}

	}

}

func (p *Plugin) openDB(file string) {

	start := time.Now()

	err := p.readDB(file)

	if err != nil {

		log.Errorf("%T load file '%s' failed, err=%s", p, file, err)

	} else {

		log.Infof("%T load file '%s' successfull, took %s", p, file, time.Since(start))

		if err = watcher.Add(file, p.openDB); err != nil {
			log.Errorf("%T watch file '%s' failed, err=%s", p, file, err)
		}

	}

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		ip := data.GetDestination()
		if !ids.IsPublicIP(ip) {
			ip = data.GetSource()
			if !ids.IsPublicIP(ip) {
				return nil
			}
		}

		if asn, exist := p.getASN(ip); exist {
			data.ASN.Number = asn.Number
			data.ASN.CountryCode = asn.CountryCode
			data.ASN.Description = asn.Description
		}

	}

	return nil
}
