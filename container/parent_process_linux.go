//go:build linux
// +build linux

package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

/*
使用exec创建一个进程，执行init命令和用户指定的初始化程序
/proc/self/exe是一个符号链接，指的是当前程序，比如ssh连接服务器时，exe->/usr/bin/bash
*/
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
	read, write, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	// 执行mini-docker init command...
	cmd := exec.Command("/proc/self/exe", "init")
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
	cmd.ExtraFiles = []*os.File{read}
	mntURL := "/home/joe/merged/"
	rootURL := "/home/joe/"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, write
}

func NewPipe() (read, write *os.File, err error) {
	read, write, err = os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

// NewWorkSpace Create an Overlay2 filesystem as container root workspace
func NewWorkSpace(rootPath string, mntURL string, volume string) {
	createLower(rootPath)
	createDirs(rootPath)
	mountOverlayFS(rootPath, mntURL)

	// 如果指定了volume则还需要mount volume
	if volume != "" {
		mntPath := path.Join(rootPath, "merged")
		hostPath, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// volumeExtract 通过冒号分割解析volume目录，比如 -v /tmp:/tmp
func volumeExtract(volume string) (sourcePath, destinationPath string, err error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid volume [%s], must split by `:`", volume)
	}

	sourcePath, destinationPath = parts[0], parts[1]
	if sourcePath == "" || destinationPath == "" {
		return "", "", fmt.Errorf("invalid volume [%s], path can't be empty", volume)
	}

	return sourcePath, destinationPath, nil
}

// mountVolume 使用 bind mount 挂载 volume
func mountVolume(mntPath, hostPath, containerPath string) {
	// 创建宿主机目录
	if err := os.Mkdir(hostPath, 0777); err != nil {
		log.Infof("mkdir parent dir %s error. %v", hostPath, err)
	}
	// 拼接出对应的容器目录在宿主机上的的位置，并创建对应目录
	containerPathInHost := path.Join(mntPath, containerPath)
	if err := os.Mkdir(containerPathInHost, 0777); err != nil {
		log.Infof("mkdir container dir %s error. %v", containerPathInHost, err)
	}
	// 通过bind mount 将宿主机目录挂载到容器目录
	// mount -o bind /hostPath /containerPath
	cmd := exec.Command("mount", "-o", "bind", hostPath, containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume failed. %v", err)
	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(rootURL string) {
	// 把busybox作为overlayfs中的lower层
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	// 检查是否已经存在busybox文件夹
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	// 不存在则创建目录并将busybox.tar解压到busybox文件夹中
	if !exist {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", busyboxURL, err)
		}
	}
}

// createDirs 创建overlayfs需要的的upper、worker目录
func createDirs(rootURL string) {
	upperURL := rootURL + "upper/"
	if err := os.Mkdir(upperURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", upperURL, err)
	}
	workURL := rootURL + "work/"
	if err := os.Mkdir(workURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", workURL, err)
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(rootURL string, mntURL string) {
	// mount -t overlay overlay -o lowerdir=lower1:lower2:lower3,upperdir=upper,workdir=work merged
	// 创建对应的挂载目录
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	// 拼接参数
	// e.g. lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/merged
	dirs := "lowerdir=" + rootURL + "busybox" + ",upperdir=" + rootURL + "upper" + ",workdir=" + rootURL + "work"
	// dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	mntPath := path.Join(rootURL, "merged")

	// 如果指定了volume则需要umount volume
	// NOTE: 一定要要先 umount volume ，然后再删除目录，否则由于 bind mount 存在，删除临时目录会导致 volume 目录中的数据丢失。
	if volume != "" {
		_, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		umountVolume(mntPath, containerPath)
	}

	umountOverlayFS(mntURL)
	deleteDirs(rootURL)
}

func umountVolume(mntPath, containerPath string) {
	// mntPath 为容器在宿主机上的挂载点，例如 /root/merged
	// containerPath 为 volume 在容器中对应的目录，例如 /root/tmp
	// containerPathInHost 则是容器中目录在宿主机上的具体位置，例如 /root/merged/root/tmp
	containerPathInHost := path.Join(mntPath, containerPath)
	cmd := exec.Command("umount", containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume failed. %v", err)
	}
}

func umountOverlayFS(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func deleteDirs(rootURL string) {
	writeURL := rootURL + "upper/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
	workURL := rootURL + "work"
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("Remove dir %s error %v", workURL, err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
