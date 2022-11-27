package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/kernelschmelze/pkg/atom"
	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/alias"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/slack-go/slack"
)

type config struct {
	Token   string
	Channel string
	Layout  layout
}

type layout struct {
	Title alias.Layout
	Text  alias.Layout
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
		plugin.NewPluginWithPriority(900),
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

	var channel string
	api := (*slack.Client)(nil)
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

			api = nil
			p.customer.Set(false)

			// re-configure

			channel = strings.TrimSpace(config.Channel)
			channel = strings.TrimLeft(channel, "#")

			if len(channel) == 0 || len(config.Token) == 0 {
				continue
			}

			if config.Layout.Title.IsValid() && config.Layout.Text.IsValid() {
				customer = config.Layout
				p.customer.Set(true)
			}

			api = slack.New(config.Token)

		case data := <-p.alias:

			if api == nil {
				continue
			}

			title := fmt.Sprint(alias.Format(data, customer.Title.Format, customer.Title.Fields...))
			text := fmt.Sprint(alias.Format(data, customer.Text.Format, customer.Text.Fields...))

			attachment := slack.Attachment{
				Title:   title,
				Text:    text,
				Pretext: data.Time,
			}

			var icon, msg string

			icon = "grey_exclamation"

			msg = data.SIDMsg

			if len(icon) > 0 {
				msg = ":" + icon + ": " + msg
			}

			_, _, err := api.PostMessage("#"+channel,
				slack.MsgOptionText(msg, false),
				slack.MsgOptionAttachments(attachment),
			)

			if err != nil {
				log.Errorf("%T post message failed, err=%s", p, err)
			}

		case data := <-p.data:

			if api == nil {
				continue
			}

			timestamp := data.GetTime()

			attachment := slack.Attachment{

				Title: fmt.Sprintf("%s %s -> %s - %d %s %s",

					data.GetProtocol(),
					data.GetSource(),
					data.GetDestination(),
					data.ASN.Number,
					data.ASN.CountryCode,
					data.ASN.Description,
				),

				Text: fmt.Sprintf("%s\n\ngid: %d sid: %d priority: %d impact: %d",

					data.SIDMap.Classification,

					data.GetGID(),
					data.GetSID(),

					data.GetPriority(),
					data.GetImpact(),
				),

				Pretext: fmt.Sprintf("%s",
					timestamp.Format("2006-01-02 15:04:05.000000"),
				),
			}

			var icon, msg string

			icon = "grey_exclamation"

			msg = data.SIDMap.Msg

			if len(icon) > 0 {
				msg = ":" + icon + ": " + msg
			}

			_, _, err := api.PostMessage("#"+channel,
				slack.MsgOptionText(msg, false),
				slack.MsgOptionAttachments(attachment),
			)

			if err != nil {
				log.Errorf("%T post message failed, err=%s", p, err)
			}

		}

	}

}
