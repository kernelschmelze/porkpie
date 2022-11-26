package pushover

import (
	"fmt"
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
	data  chan *ids.Record
	setup chan *config
	kill  chan struct{}
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(902),
		make(chan *ids.Record, 50),
		make(chan *config, 10),
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
		p.setup <- config
	}

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		// ignore internal packet

		if data.IsLocal() {
			return nil
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

		case config := <-p.setup:

			if config == nil {
				continue
			}

			app = nil
			recipient = nil

			if len(config.App) == 0 || len(config.User) == 0 {
				continue
			}

			app = po.New(config.App)
			recipient = po.NewRecipient(config.User)

			go func(app *po.Pushover, recipient *po.Recipient) {

				if app == nil || recipient == nil {
					return
				}

				if _, err := app.GetRecipientDetails(recipient); err != nil {
					log.Errorf("%T get recipient details failed, err=%s", p, err)
				}

			}(app, recipient)

		case data := <-p.data:

			if app == nil || recipient == nil {
				continue
			}

			msg := fmt.Sprintf("%s\n\n%s %s -> %s\n%d %s %s\n\ngid: %d sid: %d priority: %d impact: %d",
				data.SIDMap.Msg,
				data.GetProtocol(),
				data.GetSource(),
				data.GetDestination(),
				data.ASN.Number,
				data.ASN.CountryCode,
				data.ASN.Description,
				data.GetGID(),
				data.GetSID(),
				data.GetPriority(),
				data.GetImpact(),
			)

			title := data.SIDMap.Classification
			if len(title) > po.MessageTitleMaxLength {
				title = title[:po.MessageTitleMaxLength]
			}

			message := po.NewMessageWithTitle(msg, title)

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
