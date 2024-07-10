package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"mini-docker/cgroups/subsystems"
	"mini-docker/container"
)

// 定义run命令
var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit.
			mini-docker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "memory",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpu share limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpu set limit",
		},
	},
	/*
		这里是mini-docker执行run命令时执行的函数
		1. 判断参数是否包含command
		2. 获取用户指定的command
		3. 调用Run function 启动容器
	*/
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		cmdArray := ctx.Args()
		tty := ctx.Bool("ti")
		resourceConfig := subsystems.ResourceConfig{
			MemoryLimit: ctx.String("memory"),
			CpuSet:      ctx.String("cpuset"),
			CpuShare:    ctx.String("cpushare"),
		}
		Run(tty, cmdArray, resourceConfig)
		return nil
	},
}

// 定义了init命令，该命令是内部方法，禁止外部调用
var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside.",
	/**
	1. 获取传递过来的command参数
	2. 执行初始化操作：执行用户指定的程序
	*/
	Action: func(ctx *cli.Context) error {
		log.Infof("init container")
		cmd := ctx.Args().Get(0)

		log.Infof("command %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}
