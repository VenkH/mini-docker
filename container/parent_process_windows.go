//go:build windows
// +build windows

package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

// NewParentProcess
// 因为exec库在windows和linux环境是不同的，我们要用到linux库的实现
// 此处仅为占位，兼容的写法/*
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	return nil, nil
}
func DeleteWorkSpace(rootURL string, mntURL string) {
	log.Errorf("Withoud window implement")
}
