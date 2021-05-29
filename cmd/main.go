package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	sigKillChan := make(chan os.Signal, 1)

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Printf("crashed with: %v", r)
	// 	}
	// }()

	app := cli.App{
		Name:    "hpb-monitor",
		Usage:   "Monitor the HPB block chain find who did not produce block",
		Version: "0.0.1",
		Commands: []cli.Command{
			StartCommand(sigKillChan),
			LoopCommand(sigKillChan),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("unexpected error: ", err)
	}

}
