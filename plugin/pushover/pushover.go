package pushover

import (
	"fmt"
	"time"

	"github.com/kernelschmelze/pkg/atom"
	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/alias"
	"github.com/kernelschmelze/porkpie/ids"

	po "github.com/gregdel/pushover"
	log "github.com/kernelschmelze/pkg/logger"
)

type config struct {
	User   string
	App    string
	Layout layout
}

type layout struct {
	Title   alias.Layout
	Message alias.Layout
}

type Plugin struct {
	*plugin.PluginBase
	customer atom.Bool
	alias    chan alias.Message
	data     chan *ids.Record
	setup    chan *config
	kill     chan struct{}
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(902),
		atom.Bool{},
		make(chan alias.Message, 50),
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

	case alias.Message:

		if data.Local {
			return nil
		}

		if len(data.SIDMsg) == 0 {
			return nil
		}

		if !p.customer.IsSet() {
			return nil
		}

		select {
		case p.alias <- data:
		case <-time.After(250 * time.Millisecond):
			log.Errorf("%T alias channel full, drop message", p)
		}

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		if data.IsLocal() {
			return nil
		}

		if len(data.SIDMap.Msg) == 0 {
			return nil
		}

		if p.customer.IsSet() {
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
	customer := layout{}

	for {

		select {

		case <-p.kill:
			return

		case config := <-p.setup:

			if config == nil {
				continue
			}

			// reset

			app = nil
			recipient = nil
			customer = layout{}
			p.customer.Set(false)

			if len(config.App) == 0 || len(config.User) == 0 {
				continue
			}

			// re-configure

			app = po.New(config.App)
			recipient = po.NewRecipient(config.User)

			if config.Layout.Message.IsValid() {
				customer = config.Layout
				p.customer.Set(true)
			}

			go func(app *po.Pushover, recipient *po.Recipient) {

				if app == nil || recipient == nil {
					return
				}

				if _, err := app.GetRecipientDetails(recipient); err != nil {
					log.Errorf("%T get recipient details failed, err=%s", p, err)
				}

			}(app, recipient)

		case data := <-p.alias:

			if app == nil || recipient == nil {
				continue
			}

			msg := fmt.Sprint(alias.Format(data, customer.Message.Format, customer.Message.Fields...))

			title := data.SIDClassification

			if len(customer.Title.Format) > 0 && len(customer.Title.Fields) > 0 {
				title = fmt.Sprint(alias.Format(data, customer.Title.Format, customer.Title.Fields...))
			}

			err := p.send(app, recipient, title, msg, data.Timestamp, data.Impact)
			if err != nil {
				log.Errorf("%T send message failed, err=%s", p, err)
			}

		case data := <-p.data:

			if app == nil || recipient == nil {
				continue
			}

			msg := fmt.Sprintf("%s\n\n%s %s -> %s\n%s %s %d\n\ngid: %d sid: %d priority: %d impact: %d",
				data.SIDMap.Msg,
				data.GetProtocol(),
				data.GetSource(),
				data.GetDestination(),
				data.ASN.CountryCode,
				data.ASN.Description,
				data.ASN.Number,
				data.GetGID(),
				data.GetSID(),
				data.GetPriority(),
				data.GetImpact(),
			)

			title := data.SIDMap.Classification
			if len(title) > po.MessageTitleMaxLength {
				title = title[:po.MessageTitleMaxLength]
			}

			err := p.send(app, recipient, title, msg, data.GetTime().Unix(), data.GetImpact())
			if err != nil {
				log.Errorf("%T send message failed, err=%s", p, err)
			}

		}

	}

}

func (p *Plugin) send(app *po.Pushover, recipient *po.Recipient, title string, msg string, timestamp int64, priority uint8) error {

	if app == nil || recipient == nil {
		return fmt.Errorf("missing pushover object")
	}

	if len(title) > po.MessageTitleMaxLength {
		title = title[:po.MessageTitleMaxLength]
	}

	message := po.NewMessageWithTitle(msg, title)
	message.Timestamp = timestamp

	if priority > 0 {
		message.Priority = 1
	}

	_, err := app.SendMessage(message, recipient)

	return err
}
