// Package sysinfo collects host resource metrics.
package sysinfo

import (
	"runtime"
	"time"

	"github.com/managerhub/managerhub/shared/protocol"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// Collector gathers metrics snapshots.
type Collector struct {
	start time.Time
}

// NewCollector creates a collector tracking process uptime.
func NewCollector() *Collector { return &Collector{start: time.Now()} }

// Uptime returns agent uptime in seconds.
func (c *Collector) Uptime() int64 { return int64(time.Since(c.start).Seconds()) }

// Snapshot returns a full metrics payload.
func (c *Collector) Snapshot() protocol.Metrics {
	m := protocol.Metrics{}
	if pcts, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(pcts) > 0 {
		m.CPUPercent = pcts[0]
	}
	if cores, err := cpu.Counts(true); err == nil {
		m.CPUCores = cores
	} else {
		m.CPUCores = runtime.NumCPU()
	}
	if avg, err := load.Avg(); err == nil {
		m.Load1 = avg.Load1
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		m.RAMUsed = vm.Used
		m.RAMTotal = vm.Total
	}
	if du, err := disk.Usage("/"); err == nil {
		m.DiskUsed = du.Used
		m.DiskTotal = du.Total
	} else if du, err := disk.Usage("C:\\"); err == nil {
		m.DiskUsed = du.Used
		m.DiskTotal = du.Total
	}
	if hi, err := host.Info(); err == nil {
		m.UptimeSec = int64(hi.Uptime)
	} else {
		m.UptimeSec = c.Uptime()
	}
	if n, err := process.Pids(); err == nil {
		m.Processes = len(n)
	}
	m.GPUs = detectGPUs()
	return m
}

// HostInfo returns hostname, OS and arch.
func HostInfo() (hostname, goos, arch string) {
	return hostName(), runtime.GOOS, runtime.GOARCH
}

func hostName() string {
	if hi, err := host.Info(); err == nil {
		return hi.Hostname
	}
	return "unknown"
}

func detectGPUs() []protocol.GPUInfo {
	return detectGPUsPlatform()
}
