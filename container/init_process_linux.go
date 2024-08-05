//go:build linux
// +build linux

package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		log.Errorf("Mount error %v", err)
	}

	if err := setUpMount(); err != nil {
		log.Errorf("setup mount error %v", err)
	}
	//if err := syscall.Exec("/bin/sh", []string{"-c", "/etc/profile"}, os.Environ()); err != nil {
	//	log.Errorf("suorce error %v", err)
	//}
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

/*
*
Init 挂载点
*/
func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return err
	}
	log.Infof("Current location is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		return err
	}

	//// 构建要在 shell 中运行的命令
	//cmd := exec.Command("/bin/sh", "-c", ". /etc/profile && echo $PATH")
	//
	//// 获取命令的输出
	//if output, err := cmd.Output(); err != nil {
	//	fmt.Printf("Error: %v\n", err)
	//} else {
	//	fmt.Printf("Output: %s\n", output)
	//}

	//cmd := exec.Command("/bin/sh", "-c", "source /etc/profile")
	//if err := cmd.Run(); err != nil {
	//	log.Errorf("source failed %v", err)
	//}

	//mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		return err
	}

	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return err
	}
	return nil
}

func pivotRoot(root string) error {
	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			return err
		}
		log.Infof("file %v already exists", pivotDir)
	}
	// pivot_root 到新的rootfs, 现在老的 old_root 是挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
