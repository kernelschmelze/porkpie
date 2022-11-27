package mail

import (
	"fmt"
	"sync"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/alias"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"
)

type config struct {
	Server string
	From   Addr
	To     Addr
	Layout layout
}

type layout struct {
	Subject alias.Layout
	Body    alias.Layout
}

type Addr struct {
	Name    string
	Address string
}

type Plugin struct {
	*plugin.PluginBase
	config config
	guard  sync.RWMutex
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(901),
		config{},
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

	if c, ok := v.(*config); ok {

		p.guard.Lock()

		p.config = config{
			Server: c.Server,
			From: Addr{
				Name:    c.From.Name,
				Address: c.From.Address,
			},
			To: Addr{
				Name:    c.To.Name,
				Address: c.To.Address,
			},
			Layout: c.Layout,
		}

		p.guard.Unlock()

	}

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case alias.Message:

		p.guard.RLock()
		config := p.config
		p.guard.RUnlock()

		if len(config.Server) == 0 || len(config.From.Name) == 0 || len(config.From.Address) == 0 || len(config.To.Name) == 0 || len(config.To.Address) == 0 {
			return nil
		}

		if !config.Layout.Body.IsValid() {
			return nil
		}

		body := fmt.Sprint(alias.Format(data, config.Layout.Body.Format, config.Layout.Body.Fields...))

		subject := fmt.Sprintf("[%d] %s",
			data.Impact,
			data.SIDMsg,
		)

		if config.Layout.Subject.IsValid() {
			subject = fmt.Sprint(alias.Format(data, config.Layout.Subject.Format, config.Layout.Subject.Fields...))
		}

		from := Address(config.From.Name, config.From.Address)
		to := Address(config.To.Name, config.To.Address)

		if err := SendMail(config.Server, from, subject, body, []string{to}); err != nil {
			log.Errorf("%T send mail failed, err=%s", p, err)
		}

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		p.guard.RLock()
		config := p.config
		p.guard.RUnlock()

		if len(config.Server) == 0 || len(config.From.Name) == 0 || len(config.From.Address) == 0 || len(config.To.Name) == 0 || len(config.To.Address) == 0 {
			return nil
		}

		if config.Layout.Body.IsValid() {
			return nil
		}

		timestamp := data.GetTime()

		from := Address(config.From.Name, config.From.Address)
		to := Address(config.To.Name, config.To.Address)

		subject := fmt.Sprintf("[%d] %s",
			data.GetImpact(),
			data.SIDMap.Msg,
		)

		body := fmt.Sprintf("%s\n\n%s %s -> %s \n%d %s %s\n%s\ngid: %d sid: %d priority: %d impact: %d",
			timestamp.Format("2006-01-02 15:04:05.000000"),

			data.GetProtocol(),
			data.GetSource(),
			data.GetDestination(),
			data.ASN.Number,
			data.ASN.CountryCode,
			data.ASN.Description,

			data.SIDMap.Classification,

			data.GetGID(),
			data.GetSID(),

			data.GetPriority(),
			data.GetImpact(),
		)

		if err := SendMail(config.Server, from, subject, body, []string{to}); err != nil {
			log.Errorf("%T send mail failed, err=%s", p, err)
		}

	}

	return nil
}
