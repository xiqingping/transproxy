// +build linux

package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/jinzhu/configor"
	"github.com/xiqingping/transproxy"
	"github.com/xiqingping/transproxy/proxy"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
)

type Config struct {
	ProxyAddr  string `require:"true"`
	ListenAddr string `require:"true"`
	LogLevel   string `require:"true"`
	Uid        string
	Gid        string
}

var (
	logLevels = map[string]log.Level{
		"None":      log.None,
		"Emergency": log.Emergency,
		"Alert":     log.Alert,
		"Critical":  log.Critical,
		"Error":     log.Error,
		"Warning":   log.Warning,
		"Notice":    log.Notice,
		"Info":      log.Info,
		"Debug":     log.Debug,
	}
)

func main() {
	cfgFile := flag.String("config", "transproxyd.toml", "The config file")
	flag.Parse()

	os.Setenv("CONFIGOR_ENV_PREFIX", "TRANSPROXYD")

	cfg := &Config{}
	if err := configor.Load(cfg, *cfgFile); err != nil {
		fmt.Println("Load config file", *cfgFile, "error:", err)
		os.Exit(-1)
	}

	logLevel, ok := logLevels[cfg.LogLevel]
	if !ok {
		logLevel = log.Info
	}

	logger := golog.New(os.Stdout, logLevel)
	logger.Debug("Launching server ...")

	if err := dropPrivileges(cfg.Uid, cfg.Gid); err != nil {
		logger.Error("dropPrivileges error:", err)
		return
	}

	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		logger.Error("Listen error:", err)
		return
	}

	url, err := url.Parse(cfg.ProxyAddr)
	if err != nil {
		logger.Error("url.Parse(cfg.ProxyAddr) error:", err)
		return
	}

	proxyDialer, err := proxy.FromURL(url, proxy.Direct, nil)
	if err != nil {
		logger.Error("proxy.FromURL error:", err)
		return
	}

	tcpln := ln.(*net.TCPListener)
	bl := transproxy.NewBlackList()

	for {

		if conn, err := tcpln.AcceptTCP(); err != nil {
			logger.Error("Accept error:", err)
		} else {
			if sp, err := transproxy.NewSocketProxy(conn, bl, proxyDialer, logger); err != nil {
				logger.Error("NewSocketProxy error:", err)
			} else {
				logger.Debug("New Socket Proxy:", sp)
				go sp.Run()
			}
		}
	}
}
