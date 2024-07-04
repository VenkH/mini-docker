//go:build linux
// +build linux

package container

import (
	"os"
	"os/exec"
	"syscall"
)

/*
使用exec创建一个进程，执行init命令和用户指定的初始化程序
/proc/self/exe是一个符号链接，指的是当前程序，比如ssh连接服务器时，exe->/usr/bin/bash
*/
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	// 执行mini-docker init command...
	cmd := exec.Command("/proc/self/exe", args...)
	// 设置namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
