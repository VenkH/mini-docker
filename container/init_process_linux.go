//go:build linux
// +build linux

package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

/*
这是容器进程启动后执行的第一件事：
执行初始化，使用mount挂载proc文件系统
再通过exec执行用户的初始化命令
*/
func RunContainerInitProcess(cmd string, args []string) error {
	log.Infof("command %s", cmd)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{cmd}
	if err := syscall.Exec(cmd, argv, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
