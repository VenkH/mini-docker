package subsystems

// ResourceConfig 配置资源限制的结构体
type ResourceConfig struct {
	MemoryLimit string // 内存限制
	CpuShare    string // CPU时间片权重
	CpuSet      string // CPU核心数
}

// Subsystem 对于Cgroups各个子系统的抽象
// 将cgroup抽象成path，因为cgroup在hierarchy中的路径，便是虚拟文件系统重的虚拟路径
type Subsystem interface {
	// Name 返回subsystem的名字，如cpu/memory
	Name() string
	// Set 设置某个cgroup在这个subsystem中的资源限制
	Set(path string, res ResourceConfig) error
	// Apply 将进程添加到某个cgroup中
	Apply(path string, pid int) error
	// Remove 移除某个cgroup
	Remove(path string) error
}

var (
	SubSystemIns = []Subsystem{
		&CpuSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
