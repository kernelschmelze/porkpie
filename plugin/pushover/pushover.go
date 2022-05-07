package pushover

import (
	"fmt"
	"sync"
	"time"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	po "github.com/gregdel/pushover"
)

type config struct {
	User string
	App  string
}

type Plugin struct {
	*plugin.PluginBase
	app       *po.Pushover
	recipient *po.Recipient
	guard     sync.RWMutex
	data      chan *ids.Record
	setup     chan bool
	kill      chan struct{}
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(902),
		nil,
		nil,
		sync.RWMutex{},
		make(chan *ids.Record, 50),
		make(chan bool, 10),
		make(chan struct{}),
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

	go p.run()

	return nil
}

func (p *Plugin) stop() error {

	log.Infof("stop %T", p)

	close(p.kill)

	return nil
}

func (p *Plugin) configure(v interface{}) {

	if config, ok := v.(*config); ok {

		p.guard.Lock()

		p.app = nil
		p.recipient = nil

		if len(config.App) > 0 && len(config.User) > 0 {
			p.app = po.New(config.App)
			p.recipient = po.NewRecipient(config.User)
		}

		p.guard.Unlock()

		p.setup <- true

	}

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		// ignore internal packet
		ip := data.GetDestination()
		if !ids.IsPublicIP(ip) {
			ip = data.GetSource()
			if !ids.IsPublicIP(ip) {
				return nil
			}
		}

		if len(data.SIDMap.Msg) == 0 {
			return nil
		}

		select {
		case p.data <- data:
		case <-time.After(250 * time.Millisecond):
			log.Errorf("%T data channel full, drop message", p)
		}

	}

	return nil
}

func (p *Plugin) run() {

	app := (*po.Pushover)(nil)
	recipient := (*po.Recipient)(nil)

	for {

		select {

		case <-p.kill:
			return

		case <-p.setup:

			p.guard.RLock()

			app = p.app
			recipient = p.recipient

			p.guard.RUnlock()

		case data := <-p.data:

			if app == nil || recipient == nil {
				continue
			}

			msg := fmt.Sprintf("%s\n\n%s %s -> %s   %s %s\n\ngid: %d sid: %d priority: %d impact: %d",
				data.SIDMap.Msg,
				data.GetProtocol(),
				data.GetSource(),
				data.GetDestination(),
				data.Country,
				data.City,
				data.GetGID(),
				data.GetSID(),
				data.GetPriority(),
				data.GetImpact(),
			)

			message := po.NewMessageWithTitle(msg, data.SIDMap.Classification)

			message.Timestamp = data.GetTime().Unix()

			if data.GetImpact() > 0 {
				message.Priority = 1
			}

			_, err := app.SendMessage(message, recipient)

			if err != nil {
				log.Errorf("%T post message failed, err=%s", p, err)
			}

		}

	}

}
