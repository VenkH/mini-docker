package main

import (
	log "github.com/sirupsen/logrus"
	"mini-docker/cgroups"
	"mini-docker/cgroups/subsystems"
	"mini-docker/container"
	"os"
	"strings"
)

func Run(tty bool, cmdArray []string, res subsystems.ResourceConfig) {
	parent, write := container.NewParentProcess(tty)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	sendInitCommand(cmdArray, write)

	manager := cgroups.NewCgroupManager("mini-docker-cgroup", res)
	manager.Set()
	manager.Apply(parent.Process.Pid)
	defer manager.Destroy()

	parent.Wait()
	os.Exit(-1)
}
func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is %v", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
