package subsystems

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(cgroupPath string, res ResourceConfig) error {
	if subSystemGroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuShare != "" {
			if err := os.WriteFile(path.Join(subSystemGroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu share fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subSystemGroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subSystemGroupPath)
	} else {
		return err
	}
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subSystemGroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subSystemGroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
