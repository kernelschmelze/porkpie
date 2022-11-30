package logger

import (
	"fmt"
	"os"
	"sync"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/alias"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	"golang.org/x/crypto/ssh/terminal"
)

type config struct {
	Layout alias.Layout
}

type Plugin struct {
	*plugin.PluginBase
	guard  sync.RWMutex
	config config
	isTTY  bool
}

func init() {
	New()
}

func New() *Plugin {

	isTTY := true
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		isTTY = false
	}

	p := &Plugin{
		plugin.NewPluginWithPriority(999),
		sync.RWMutex{},
		config{},
		isTTY,
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
			Layout: c.Layout,
		}
		p.guard.Unlock()

	}

}

func (p *Plugin) getLayout() alias.Layout {

	p.guard.RLock()
	layout := p.config.Layout
	p.guard.RUnlock()

	return layout
}

func (p *Plugin) do(v interface{}) error {

	if !p.isTTY {
		return nil
	}

	switch data := v.(type) {

	case alias.Message:

		layout := p.getLayout()

		if !layout.IsValid() {
			return nil // we have no valid customer layout, skip
		}

		// if data.Local {
		// 	return nil
		// }

		fmt.Println(alias.Format(data, layout.Format, layout.Fields...))

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		layout := p.getLayout()

		if layout.IsValid() { // we have a customer layout, skip default logging
			return nil
		}

		timestamp := data.GetTime()

		fmt.Printf("%s - sid:%d gid:%d - %s %s -> %s - [%d] %s - %s (%d %s %s)\n",
			timestamp.Format("2006-01-02 15:04:05.000000"),
			data.GetSID(),
			data.GetGID(),

			data.GetProtocol(),
			data.GetSource(),
			data.GetDestination(),

			data.GetPriority(),
			data.SIDMap.Classification,

			data.SIDMap.Msg,

			data.ASN.Number,
			data.ASN.CountryCode,
			data.ASN.Description,
		)

	}

	return nil
}
