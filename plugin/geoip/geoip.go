package geoip

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/pkg/plugin/watcher"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/oschwald/geoip2-golang"
)

type config struct {
	GeoDB string
}

type Plugin struct {
	*plugin.PluginBase
	db          *geoip2.Reader
	guard       sync.RWMutex
	currentFile string
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(500),
		nil,
		sync.RWMutex{},
		"",
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

	p.closeGeoIP2()

	return nil
}

func (p *Plugin) configure(v interface{}) {

	if config, ok := v.(*config); ok {

		if len(config.GeoDB) > 0 {
			p.openGeoIP2(config.GeoDB)
		} else {
			p.closeGeoIP2()
		}

	}

}

func (p *Plugin) openGeoIP2(file string) {

	p.closeGeoIP2()

	p.guard.Lock()

	var err error

	start := time.Now()

	p.db, err = geoip2.Open(file)

	if err != nil {

		log.Errorf("%T load file '%s' failed, err=%s", p, file, err)

	} else {

		log.Infof("%T load file '%s' successfull, took %s", p, file, time.Since(start))

		if err = watcher.Add(file, p.openGeoIP2); err != nil {
			log.Errorf("%T watch file '%s' failed, err=%s", p, p.currentFile, err)
		} else {
			p.currentFile = file
		}

	}

	p.guard.Unlock()

}

func (p *Plugin) closeGeoIP2() {

	p.guard.Lock()

	if p.db != nil {

		p.db.Close()
		p.db = nil

		if len(p.currentFile) > 0 {
			watcher.Remove(p.currentFile)
			p.currentFile = ""
		}

	}

	p.guard.Unlock()

}

func (p *Plugin) getCity(ip net.IP) (*geoip2.City, error) {

	p.guard.Lock()
	defer p.guard.Unlock()

	if p.db == nil {
		return nil, errors.New("geoip2 not available")
	}

	city, err := p.db.City(ip)
	return city, err

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

		if record, err := p.getCity(ip); err == nil {

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

			data.MM.Country = record.Country.IsoCode
			data.MM.City = city

		}

	}

	return nil
}
