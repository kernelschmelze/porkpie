package logger

import (
	"fmt"
	"os"

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"

	"golang.org/x/crypto/ssh/terminal"
)

type config struct {
}

type Plugin struct {
	*plugin.PluginBase
	isTTY bool
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

}

func (p *Plugin) do(v interface{}) error {

	if !p.isTTY {
		return nil
	}

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		timestamp := data.GetTime()

		fmt.Printf("%s - %d:%d \t - %s %s -> %s \t - %d %s %s %s - %s\n",
			timestamp.Format("2006-01-02 15:04:05.000000"),
			data.GetGID(),
			data.GetSID(),

			data.GetProtocol(),
			data.GetSource(),
			data.GetDestination(),

			data.GetPriority(),
			data.SIDMap.Classification,

			data.Country,
			data.City,

			data.SIDMap.Msg,
		)

	}

	return nil
}
