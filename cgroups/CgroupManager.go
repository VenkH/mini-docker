package cgroups

import (
	"github.com/sirupsen/logrus"
	"mini-docker/cgroups/subsystems"
)

type CgroupManager struct {
	// cgroup在hierarchy中的路径
	// 相当于创建的cgroup目录相对于root cgroup目录的路径
	Path string
	// 容器的资源限制，一旦创建之后不可修改
	subsystems.ResourceConfig
}

func NewCgroupManager(path string, res subsystems.ResourceConfig) *CgroupManager {
	return &CgroupManager{
		Path:           path,
		ResourceConfig: res,
	}
}

// Apply 将进程pid加入到这个cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubSystemIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// Set 设置cgroup资源限制
func (c *CgroupManager) Set() error {
	for _, subSysIns := range subsystems.SubSystemIns {
		subSysIns.Set(c.Path, c.ResourceConfig)
	}
	return nil
}

// Destroy 释放cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubSystemIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}
