package main

import (
	"fmt"
	api2 "git.diabin.com/BlockChain/qitmeer-presstest/api"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/presstest"
	"github.com/bCoder778/log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

func main() {

	run()

}

func setSystemResource(number int) {
	runtime.GOMAXPROCS(number)
}

var press *presstest.Engine
var api *api2.Api
var wg *sync.WaitGroup

func run() {
	var err error
	wg = &sync.WaitGroup{}
	if err := config.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	setSystemResource(config.Setting.Cpu.Number)
	log.SetOption(&log.Option{
		LogLevel: config.Setting.Log.Level,
		Mode:     config.Setting.Log.Mode,
		Path:     config.Setting.Log.Path,
	})

	api, err = api2.NewApi(config.Setting)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	go listenInterrupt()

	// wg.Add(1)
	api.Run()

	// wg.Wait()
	time.Sleep(time.Second * 5)
}

func listenInterrupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c

		api.Stop()

	}()
}
