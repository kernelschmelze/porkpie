package main

import (
	_ "github.com/kernelschmelze/porkpie/plugin/filter"
	_ "github.com/kernelschmelze/porkpie/plugin/geoip"
	_ "github.com/kernelschmelze/porkpie/plugin/logger"
	_ "github.com/kernelschmelze/porkpie/plugin/mail"
	_ "github.com/kernelschmelze/porkpie/plugin/pushover"
	_ "github.com/kernelschmelze/porkpie/plugin/reader"
	_ "github.com/kernelschmelze/porkpie/plugin/sidmap"
	_ "github.com/kernelschmelze/porkpie/plugin/slack"

	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kernelschmelze/pkg/plugin/config"
	"github.com/kernelschmelze/pkg/plugin/manager"

	log "github.com/kernelschmelze/pkg/logger"
)

func main() {

	path := "./config.toml"
	if err := config.Read(path); err != nil {
		log.Errorf("read config '%s' failed, err=%s", path, err)
	}

	manager := plugin.GetManager()

	defer func() {
		config.Close() // close the file watcher, it is save to call config.Write
		manager.Stop()
	}()

	manager.Start()

	signalHandler()

	go func() {
		select {
		case <-time.After(10 * time.Second):
			os.Exit(1)
		}
	}()

}

func signalHandler() {

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, os.Interrupt)

	select {
	case <-gracefulStop:
		fmt.Println("")

	}

}
