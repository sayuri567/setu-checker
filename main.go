package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/sayuri567/setu-checker/config"
	"github.com/sayuri567/setu-checker/manager"
	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config", "config.yaml", "program config")
)

func main() {
	err := config.Conf.Load(*configFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := manager.Mod.Init(); err != nil {
		panic(err)
	}
	manager.Mod.Run()

	shutdown()
}

func shutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	go func() {
		time.Sleep(30 * time.Second)
		logrus.WithField("signal", s).Warn("program waiting stop timeout")
		os.Exit(1)
	}()
	manager.Mod.Stop()
	logrus.WithField("signal", s).Info("program stopped")
	os.Exit(0)
}
