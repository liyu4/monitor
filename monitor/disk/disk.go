package disk

import (
	"fmt"
	"math"
	"time"

	"monitor/alert"
	"monitor/format"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/shirou/gopsutil/disk"
)

// Single mountedpoint disk.
type Disk struct {
	DiskTotal   uint64
	DiskUsed    uint64
	DiskFree    uint64
	DiskPercent float64
	Path        string
}

// All mountedpoint disk.
type AllDisk struct {
	Ad []Disk
}

func newAllDisk() *AllDisk {
	return &AllDisk{}
}

func GetDiskUsage() (*AllDisk, error) {
	disk := newAllDisk()
	return disk.getDiskUsage()
}

func (a *AllDisk) getDiskUsage() (*AllDisk, error) {
	mountpaths, err := getMountPath()

	if err != nil {
		return nil, err
	}
	for _, v := range mountpaths {
		disk, err := disk.Usage(v)
		if err != nil {
			logx.FError("getDiskUsage to get disk usage error: %v")
			return nil, err
		}
		var percent float64

		if math.IsNaN(disk.UsedPercent) {
			percent = 0
		} else {
			percent = disk.UsedPercent
		}

		a.Ad = append(a.Ad, Disk{
			DiskTotal:   disk.Total,
			DiskUsed:    disk.Used,
			DiskFree:    disk.Free,
			DiskPercent: percent,
			Path:        disk.Path,
		})
	}
	return a, nil
}

// get all disk useage.
func getMountPath() ([]string, error) {
	partitions, err := disk.Partitions(true)

	if err != nil {
		logx.FError("getMountPath get mmouton path error: %v", err)
		return nil, err
	}

	mountpaths := make([]string, 0)

	for _, part := range partitions {
		mountpaths = append(mountpaths, part.Mountpoint)
	}
	return mountpaths, nil
}

func MonitorAllDisk() {
	go func() {
		disk := newAllDisk()
		err := disk.monitorAllDisk()
		logx.FInfo("%v", "Disk module started!")
		if err != nil {
			logx.FCritical("MonitorAllDisk error: %v", err)
		}
	}()
}

func (a *AllDisk) monitorAllDisk() error {
	cf := parse.NewCf()

	if cf.DiskConfig == nil {
		logx.FError("%v", "DiskConfig is nil!")
		return fmt.Errorf("DiskConfig is nil!")
	}

	a.reset()
	t1 := time.NewTicker(time.Second * time.Duration(cf.DiskConfig.Lasttime))
	for {
		select {
		case <-t1.C:
			disks, err := a.getDiskUsage()

			if err != nil {
				continue
			}

			for _, v := range disks.Ad {
				logx.FDebug("Disk infos Path: %s  Toatal: %s Used: %s Free: %s Percent: %s%s ", v.Path, format.Translate(float64(v.DiskTotal)), format.Translate(float64(v.DiskUsed)), format.Translate(float64(v.DiskFree)), format.Float64(v.DiskPercent), "%")
				if v.DiskFree <= uint64(cf.DiskConfig.Free) && v.DiskTotal > uint64(cf.DiskConfig.Condition) {
					content := fmt.Sprintf("Alarm: monitor disk %s free  %d beycond the threshold! \nform: %v ", v.Path, format.Translate(float64(v.DiskUsed)), cf.Addr)
					header := "Mounted on" + v.Path + " used exceeds the threshold value!"
					timestamp := time.Now()
					alert.AlertConvergence(cf.DiskConfig.EmailArray, timestamp, content, header)
				}
			}
			a.reset()
		}
	}
	return nil
}

func (a *AllDisk) reset() {
	a.Ad = a.Ad[:0:0]
}
