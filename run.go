package main

import (
	log "github.com/sirupsen/logrus"
	"mini-docker/container"
	"os"
)

func Run(tty bool, cmd string) {
	parent := container.NewParentProcess(tty, cmd)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	parent.Wait()
	os.Exit(-1)
}
