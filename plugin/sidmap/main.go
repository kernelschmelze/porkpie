package sidmap

import (
	"sync"
	"time"

	"github.com/kernelschmelze/pkg/path"
	"github.com/kernelschmelze/pkg/plugin/plugin/base"

	log "github.com/kernelschmelze/pkg/logger"
	watcher "github.com/kernelschmelze/pkg/plugin/watcher"

	"github.com/kernelschmelze/porkpie/ids"
)

type config struct {
	Maps []string
}

type Plugin struct {
	*plugin.PluginBase
	files      []string
	filesGuard sync.RWMutex
	maps       map[string]ids.SIDMapItem
	mapsGuard  sync.RWMutex
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPlugin(),
		[]string{},
		sync.RWMutex{},
		make(map[string]ids.SIDMapItem),
		sync.RWMutex{},
	}

	err := p.Init(plugin.PluginConfig{
		p,
		p.start,
		p.stop,
		p.configure,
		p.processRecord,
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

func (p *Plugin) processFile(file string) {

	start := time.Now()

	if err := p.loadMap(file); err != nil {
		log.Errorf("%T load file '%s' failed, err=%s", p, file, err)
	} else {
		log.Infof("%T load file '%s' successfull, took %s", p, file, time.Since(start))
	}

}

func (p *Plugin) configure(v interface{}) {

	if config, ok := v.(*config); ok {

		var files []string

		for i := range config.Maps {

			file := config.Maps[i]
			file, _ = utils.ExpandPath(file)

			if len(file) > 0 {
				log.Debugf("%T use file '%s'", p, file)
				files = append(files, file)
			}

		}

		p.filesGuard.RLock()
		for i := range p.files {
			watcher.Remove(files[i])
		}
		p.filesGuard.RUnlock()

		for i := range files {

			file := files[i]
			p.processFile(file)

			if err := watcher.Add(file, func(file string) {
				// todo: throttled
				p.processFile(file)
			}); err != nil {
				log.Errorf("%T watch file '%s' failed, err=%s", p, file, err)
			}

		}

		p.filesGuard.Lock()
		p.files = files
		p.filesGuard.Unlock()

	}

}

func (p *Plugin) processRecord(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		gid := data.GetGID()
		sid := data.GetSID()

		index := p.getIndex(gid, sid)

		p.mapsGuard.RLock()
		item, exist := p.maps[index]
		p.mapsGuard.RUnlock()

		if exist {
			data.SIDMap = item
		}

	}

	return nil
}
