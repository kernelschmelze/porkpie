package geoip

import (
	"time"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/oschwald/geoip2-golang"
)

type config struct {
	GeoDB string
}

type Plugin struct {
	*plugin.PluginBase
	db *geoip2.Reader
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPlugin(),
		nil,
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

		if len(config.GeoDB) > 0 {

			var err error
			start := time.Now()

			if p.db, err = geoip2.Open(config.GeoDB); err != nil {
				log.Errorf("%T load file '%s' failed, err=%s", p, config.GeoDB, err)
			} else {
				log.Infof("%T load file '%s' successfull, took %s", p, config.GeoDB, time.Since(start))
			}

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

		if record, err := p.db.City(ip); err == nil {

			language := "de"
			city := record.City.Names[language]

			if len(city) == 0 {
				language = "en"
				city = record.City.Names[language]
			}

			if len(city) == 0 {

				for _, v := range record.City.Names {
					if len(v) > 0 {
						city = v
						break
					}
				}
			}

			data.Country = record.Country.IsoCode
			data.City = city

		}

	}

	return nil
}
