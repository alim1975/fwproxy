package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"time"

	"proxy"
)

var (
	logger = logrus.StandardLogger()
)

func main() {
	conf := flag.String("conf", "", "Proxy config file.")
	flag.Parse()
	if *conf == "" {
		logger.Error("Proxy config file is missing.")
		return
	}
	data, err := ioutil.ReadFile(*conf)
	if err != nil {
		logger.Error("Failed to read proxy config file: ", err)
		return
	}
	var cfg proxy.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.Error("Failed to parse proxy config YAML file: ", err)
		return
	}
	cli := &http.Client{Timeout: 1 * time.Second}
	handler := proxy.New(cfg.Listener, cfg.Urldb, cli)
	err = handler.Start()
	logger.Error("Proxy HTTP handler error: ", err)
}
