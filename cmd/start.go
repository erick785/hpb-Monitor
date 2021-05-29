package main

import (
	"os"

	"github.com/erick785/hpb-monitor/server"
	"github.com/urfave/cli"
)

func StartCommand(sigKillChan chan os.Signal) cli.Command {

	m := server.NewMonitor(server.DefaultConfig, sigKillChan)
	return cli.Command{
		Name:   "start",
		Usage:  "Monitor the HPB block chain find who did not produce block",
		Action: m.Start,
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
			cli.Int64Flag{
				Name:        "start-block",
				Value:       0,
				Usage:       "block number of the first block to begin scanning",
				Destination: &server.DefaultConfig.StartBlock,
			},
			cli.Int64Flag{
				Name:        "end-block",
				Value:       0,
				Usage:       "block number of the block to end scanning",
				Destination: &server.DefaultConfig.EndBlock,
			},
		},
	}
}
