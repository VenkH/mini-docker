package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = `mini-docker is a simple container runtime implementation.
				The purpose of this project is to learn how docker works and how to write a docker by ourseleves.`

func main() {
	app := cli.NewApp()
	app.Name = "mini-docker"
	app.Usage = usage

	app.Commands = []cli.Command{}
	app.Before = func(context *cli.Context) error {
		// 设置日志格式化工具
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
