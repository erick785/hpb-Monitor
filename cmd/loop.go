package main

import (
	"os"

	"github.com/erick785/hpb-monitor/server"
	"github.com/urfave/cli"
)

func LoopCommand(sigKillChan chan os.Signal) cli.Command {
	m := server.NewMonitor(server.DefaultConfig, sigKillChan)
	return cli.Command{
		Name:   "loop",
		Usage:  "Monitor the HPB block chain find who did not produce block",
		Action: m.Loop,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "hpb-host",
				Value:       "http://hpbnode.com",
				Usage:       "host endpoint of the hpb node",
				Destination: &server.DefaultConfig.NodeEndpoint,
			},
			cli.StringFlag{
				Name:        "hpb-scan-host",
				Value:       "http://hpbscan.org",
				Usage:       "host endpoint of the hpb scan",
				Destination: &server.DefaultConfig.ScanEndpoint,
			},
			cli.StringFlag{
				Name:        "monitor-web-url",
				Value:       "",
				Usage:       "URL of the hpb monitor",
				Destination: &server.DefaultConfig.HttpURL,
			},
			cli.StringFlag{
				Name:        "monitor-web-port",
				Value:       "9090",
				Usage:       "port of the hpb monitor",
				Destination: &server.DefaultConfig.HttpPort,
			},
		},
	}
}
