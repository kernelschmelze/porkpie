### porkpie

Porkpie is a snort unified log spool reader that translates snort events into human readable form and delivers them as slack, pushover and local mail messages. Configuration settings are applied in real time, changes to the sidmap file are automatically detected and re-read. There is no need to restart the program.
  

### build from source

`go get github.com/kernelschmelze/porkpie`

### setup

Make sure that the `config.toml` file is located in the working directory.
  

``` bash
gokay@mazikeen ~/porkpie
# ll                                         
total 70096
-rw-r--r-- 1 gokay gokay      539 Aug  2 10:53 config.toml
-rw-r--r-- 1 gokay gokay 62534967 Jul 26 13:23 GeoLite2-City.mmdb
-rwxr-x--x 1 gokay gokay  9234917 Aug  1 17:38 porkpie
```

`config.toml`  

``` toml

[filter]
  pattern = ["1:2525017", "1:10000001", "1:2525016"]

[geoip]
  geodb = "./GeoLite2-City.mmdb"

[ip2asn]
	db ="./ip2asn-combined.tsv.gz"

[mail]
  server = "127.0.0.1:25"

  [mail.from]
    address = "snort@localhost"
    name = "snort"

  [mail.to]
    address = "root@localhost"
    name = "root"

[reader]
  fileprefix = "snort.u2"
  path = "/var/log/snort"

[sidmap]
  maps = ["/etc/snort/sid-msg.map"]

[slack]
  channel = "ids"
  token = "<your slack token>"

[pushover]
  app = "<your API token>"
  user = "<your user key>"

```


### install as service

`porkpie.service`  


``` bash
[Unit]
Description=snort notify
Wants=network.target
After=network.target

[Service]
Type=simple
PermissionsStartOnly=true
WorkingDirectory=/home/gokay/porkpie
ExecStart=/home/gokay/porkpie/porkpie
Restart=always
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=porkpie

[Install]
WantedBy=multi-user.target
```

``` bash
sudo cp porkpie.service /lib/systemd/system/.
sudo chmod 755 /lib/systemd/system/porkpie.service
sudo systemctl enable porkpie.service
sudo systemctl start porkpie
```

### extend porkpie with your own cool plugin

``` go
package main

import (
	...

	_ "github.com/kernelschmelze/porkpie/plugin/yourplugin"	

	...
)

```

``` go
package yourplugin

import (

	"github.com/kernelschmelze/pkg/plugin/plugin/base"
	"github.com/kernelschmelze/porkpie/ids"

	log "github.com/kernelschmelze/pkg/logger"
)

type config struct {
	Key string
}

type Plugin struct {
	*plugin.PluginBase
}

func init() {
	New()
}

func New() *Plugin {

	p := &Plugin{
		plugin.NewPluginWithPriority(950),
	}

	err := p.Init(plugin.PluginConfig{
		p,
		p.start, 	 // plugin start callback
		p.stop, 	 // plugin stop callback
		p.configure, // config changed callback
		p.do,		 // do something with the snort record
		&config{},	 // your config object
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

		// do some cool stuff with your config.Key

	}

}

func (p *Plugin) do(v interface{}) error {

	switch data := v.(type) {

	case *ids.Record:

		if !data.IsValid() || data.Drop {
			return nil
		}

		// do some cool stuff with the snort record

	}

	return nil
}


```

### hints
  

Make sure that the snort output plugin for unified2 is enabled.  
  
`snort.conf`  

```conf
###################################################
# Step #6: Configure output plugins
# For more information, see Snort Manual, Configuring Snort - Output Modules
###################################################

# unified2 
# Recommended for most installs
# output unified2: filename merged.log, limit 128, nostamp, mpls_event_types, vlan_event_types
output unified2: filename snort.u2, limit 128

```  
  
 Make sure that porkpie is running in a user context that has read permissions on the snort log directory.  
