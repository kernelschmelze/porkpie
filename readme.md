### porkpie

Porkpie is a snort unified log spool reader that translates snort events into human readable form and delivers them as slack and local mail messages. Configuration settings are applied in real time, changes to the sidmap file are automatically detected and re-read. There is no need to restart the program.
  

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