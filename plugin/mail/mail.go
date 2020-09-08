package mail

import (
	"fmt"
	//"regexp"
	"sync"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"
)

type config struct {
	Server string
	From   Addr
	To     Addr
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
		}

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
		config := p.config
		p.guard.RUnlock()

		if len(config.Server) == 0 || len(config.From.Name) == 0 || len(config.From.Address) == 0 || len(config.To.Name) == 0 || len(config.To.Address) == 0 {
			return nil
		}

		timestamp := data.GetTime()

		from := Address(config.From.Name, config.From.Address)
		to := Address(config.To.Name, config.To.Address)

		subject := fmt.Sprintf("[%d] %s",
			data.GetImpact(),
			data.SIDMap.Msg,
		)

		// var payload string

		// body := fmt.Sprintf("%s\n%d:%d %s %s\n%s -> %s %s %s\n\n%s",
		// 	timestamp.Format("2006-01-02 15:04:05.000000"),
		// 	data.GetGID(), data.GetSID(), data.GetProtocol(), data.SIDMap.Classification,
		// 	data.GetSource(), data.GetDestination(), data.Country, data.City,
		// 	payload,
		// )

		body := fmt.Sprintf("%s\n\n%s %s -> %s   %s %s\n%s\ngid: %d sid: %d priority: %d impact: %d",
			timestamp.Format("2006-01-02 15:04:05.000000"),

			data.GetProtocol(),
			data.GetSource(),
			data.GetDestination(),
			data.Country,
			data.City,

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

// func (p *Plugin) getPrintable(data []byte) string {
// 	re := regexp.MustCompile("[^a-zA-Z0-9]+")
// 	str := re.ReplaceAllString(string(data[:]), ".")
// 	return str
// }
