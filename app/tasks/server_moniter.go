package tasks

import (
	"github.com/askasoft/pango/cog/ringbuffer"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/oss/cpu"
	"github.com/askasoft/pango/oss/disk"
	"github.com/askasoft/pango/oss/mem"
)

var (
	disks ringbuffer.RingBuffer[uint64]
	cpus  ringbuffer.RingBuffer[float64]
	mems  ringbuffer.RingBuffer[float64]
)

func MonitorServer() {
	monitorDisk()
	monitorCPUUsage()
	monitorMemUsage()
}

func monitorDisk() {
	diskFree := ini.GetSize("monitor", "diskFree")
	if diskFree <= 0 {
		return
	}

	du, err := disk.GetDiskUsage(".")
	if err != nil {
		log.GetLogger("MONITOR").Error(err)
		return
	}

	disks.Push(du.Available)

	diskCount := ini.GetInt("monitor", "diskCount", 1)
	if disks.Len() > diskCount {
		disks.Poll()
	}

	if disks.Len() == diskCount {
		daa := calcAverage(disks)
		if daa < uint64(diskFree) {
			log.GetLogger("MONITOR").Errorf("insufficient free disk space %s", num.HumanSize(du.Available))
			disks.Clear()
		}
	}
}

var lastCPUStats cpu.CPUStats

func monitorCPUUsage() {
	cpuUsage := ini.GetFloat("monitor", "cpuUsage")
	if cpuUsage <= 0 {
		return
	}

	cs, err := cpu.GetCPUStats()
	if err != nil {
		log.GetLogger("MONITOR").Error(err)
		return
	}

	thisCPUUsage := cs
	cs.Subtract(&lastCPUStats)
	lastCPUStats = thisCPUUsage

	cpus.Push(cs.CPUUsage())

	cpuCount := ini.GetInt("monitor", "cpuCount", 1)
	if cpus.Len() > cpuCount {
		cpus.Poll()
	}

	if cpus.Len() == cpuCount {
		cua := calcAverage(cpus)
		if cua > cpuUsage {
			log.GetLogger("MONITOR").Errorf("cpu usage %.2f%% is too high", cua*100)
			cpus.Clear()
		}
	}
}

func monitorMemUsage() {
	memUsage := ini.GetFloat("monitor", "memUsage")
	if memUsage <= 0 {
		return
	}

	ms, err := mem.GetMemoryStats()
	if err != nil {
		log.GetLogger("MONITOR").Error(err)
		return
	}

	mems.Push(ms.Usage())

	memCount := ini.GetInt("monitor", "memCount", 1)
	if mems.Len() > memCount {
		mems.Poll()
	}

	if mems.Len() == memCount {
		mua := calcAverage(mems)
		if mua > memUsage {
			log.GetLogger("MONITOR").Errorf("memory usage %.2f%% is too high", mua*100)
			mems.Clear()
		}
	}
}

func calcAverage[E uint64 | float64](rb ringbuffer.RingBuffer[E]) E {
	var total E
	for it := rb.Iterator(); it.Next(); {
		total += it.Value()
	}
	return total / E(rb.Len())
}
