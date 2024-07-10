//go:build linux
// +build linux

package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

/*
这是容器进程启动后执行的第一件事：
执行初始化，使用mount挂载proc文件系统
再通过exec执行用户的初始化命令
*/
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}
	log.Infof("command %v", cmdArray)

	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示
	// 声明你要这个新的mount namespace独立。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec look path error %v", err)
		return err
	}
	log.Infof("Find path %v", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
func readUserCommand() []string {
	// 获取当前进程的第4个文件描述符，匿名管道的read一端
	// 一个进程默认有3个文件描述符，分别是标准输入，标准输出，标准错误
	// 此时是在父进程中给子进程添加了匿名管道的文件描述符
	readPipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
