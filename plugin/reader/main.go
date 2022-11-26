package reader

import (
	"sync"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"

	log "github.com/kernelschmelze/pkg/logger"
	store "github.com/kernelschmelze/pkg/plugin/config"
)

type config struct {
	Path        string
	FilePrefix  string
	CurrentFile string
	Offset      int64
}

type bookmark struct {
	File    string
	Offset  int64
	Changed bool
}

type Plugin struct {
	*plugin.PluginBase
	config     config
	bookmark   bookmark
	jobs       chan job
	kill       chan bool
	killReader chan bool
	wgJobs     sync.WaitGroup
	wgReader   sync.WaitGroup
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(1000),
		config{},
		bookmark{},
		make(chan job, 1024),
		nil,
		nil,
		sync.WaitGroup{},
		sync.WaitGroup{},
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

	p.kill = make(chan bool)
	p.killReader = make(chan bool)

	p.wgJobs.Add(1)
	go p.handleJobs()

	p.wgReader.Add(1)
	go p.read(p.config)

	return nil
}

func (p *Plugin) stop() error {

	log.Infof("stop %T", p)

	close(p.killReader)
	p.wgReader.Wait()

	close(p.kill)
	p.wgJobs.Wait()

	p.storeBookmark()

	return nil
}

func (p *Plugin) restart(c config) {

	close(p.killReader)
	p.wgReader.Wait()

	p.killReader = make(chan bool)

	p.config = c

	p.wgReader.Add(1)
	go p.read(p.config)

}

func (p *Plugin) configure(v interface{}) {

	if config, ok := v.(*config); ok {

		if p.IsActivated() {

			if p.config.Path != config.Path || p.config.FilePrefix != config.FilePrefix {
				p.restart(*config)
				return
			}

		}

		p.config = *config

	} else {
		log.Debugf("unknown config object %T", v)
	}

}

func (p *Plugin) setBookmark(file string, offset int64) {

	if p.bookmark.File == file && p.bookmark.Offset == offset {
		return
	}

	p.bookmark.File = file
	p.bookmark.Offset = offset
	p.bookmark.Changed = true
}

func (p *Plugin) storeBookmark() {

	return // DEBUG

	if !p.bookmark.Changed {
		return
	}

	store.Write(p, "CurrentFile", p.bookmark.File)
	store.Write(p, "Offset", p.bookmark.Offset)

	p.bookmark.Changed = false
}
