package main

import (
	log "github.com/sirupsen/logrus"
	"mini-docker/cgroups"
	"mini-docker/cgroups/subsystems"
	"mini-docker/container"
	"os"
)

func Run(tty bool, cmd string, res subsystems.ResourceConfig) {
	parent := container.NewParentProcess(tty, cmd)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	manager := cgroups.NewCgroupManager("mini-docker-cgroup", res)
	manager.Set()
	manager.Apply(parent.Process.Pid)
	defer manager.Destroy()

	parent.Wait()
	os.Exit(-1)
}
