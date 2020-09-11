package slack

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/slack-go/slack"
)

type config struct {
	Token   string
	Channel string
}

type Plugin struct {
	*plugin.PluginBase
	api     *slack.Client
	channel string
	guard   sync.RWMutex
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(900),
		nil,
		"",
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

		channel := config.Channel
		channel = strings.TrimSpace(channel)
		channel = strings.TrimLeft(channel, "#")

		if p.api == nil && (len(channel) == 0 || len(config.Token) == 0) {
			return
		}

		p.guard.Lock()
		p.api = slack.New(config.Token)
		p.channel = channel
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
		channel := p.channel
		api := p.api
		p.guard.RUnlock()

		if api == nil || len(channel) == 0 {
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

		timestamp := data.GetTime()

		attachment := slack.Attachment{

			Title: fmt.Sprintf("%s %s -> %s   %s %s",

				data.GetProtocol(),
				data.GetSource(),
				data.GetDestination(),
				data.Country,
				data.City,
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
		if len(msg) == 0 {
			return nil
			msg = "unknown alert"
		}

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

	return nil
}
