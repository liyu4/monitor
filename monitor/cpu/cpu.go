package cpu

import (
	"fmt"
	"sync"
	"time"

	"monitor/alert"
	"monitor/format"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/shirou/gopsutil/cpu"
)

type CpuUseage struct {
	CurrentUsage float64
	CpuMax       float64
	CpuMin       float64
	CpuAver      float64
}

var (
	record *CpuUseage
	lock   = sync.RWMutex{}
)

func NewCpu() *CpuUseage {
	return &CpuUseage{}
}

// Five hundred milliseconds check cpu useage.
func (c *CpuUseage) GetCpuUsage() (currentuseage *CpuUseage, err error) {

	percent, err := cpu.Percent(time.Millisecond*500, false)
	if err != nil {
		logx.FError("GetCpuUsage to get cpu usage error: %v", err)
		return &CpuUseage{}, err
	}

	if len(percent) == 0 {
		logx.FError("GetCpuUsage percent is nil")
		return &CpuUseage{}, fmt.Errorf("GetCpuUsage percent is nil")
	}

	//Init
	if c.CurrentUsage == 0 && c.CpuMin == 0 && c.CpuAver == 0 && c.CpuMax == 0 {
		c.CurrentUsage = percent[0]
		c.CpuMax = percent[0]
		c.CpuMin = percent[0]
		c.CpuAver = percent[0]
	}

	if percent[0] > c.CpuMax {
		c.CpuMax = percent[0]
		c.CurrentUsage = percent[0]
	} else {
		if c.CpuMin > percent[0] {
			c.CpuMin = percent[0]
		}
		c.CurrentUsage = percent[0]
	}

	if c.CpuAver != percent[0] {
		c.CpuAver = (c.CpuAver + percent[0]) / 2
	}
	return c, nil
}

func MonitorCpu() {
	go func() {
		cpu := NewCpu()
		logx.FInfo("%v", "Cpu module started!")
		err := cpu.monitorCpu()
		if err != nil {
			logx.FCritical("MonitorCpu error: %v", err)
		}
	}()
}

func Result() *CpuUseage {
	return record
}

func (c *CpuUseage) Clear() {
	c.CurrentUsage = 0
	c.CpuMax = 0
	c.CpuMin = 0
	c.CpuAver = 0

}

func (c *CpuUseage) monitorCpu() error {

	cf := parse.NewCf()

	if cf.CpuConfig == nil {
		logx.FError("%v", "CpuConfig is nil!")
		return fmt.Errorf("CpuConfig is nil!")
	}

	t1 := time.NewTicker(time.Millisecond * time.Duration(cf.CpuConfig.Duration))

	hasRetry := false
	count := (cf.CpuConfig.Lasttime * (1000)) / cf.CpuConfig.Duration
	startCount := 0
	for {
		select {
		case <-t1.C:
			cpus, err := c.GetCpuUsage()
			logx.FDebug("cpu usage: Max: %v%s Min: %v%s Average: %v%s Current: %v%s ", format.Float64(cpus.CpuMax), "%", format.Float64(cpus.CpuMin), "%", format.Float64(cpus.CpuAver), "%", format.Float64(cpus.CurrentUsage), "%")
			if err != nil {
				return err
			}

			lock.Lock()
			record = c
			lock.Unlock()

			// For test.
			if cpus.CurrentUsage-cf.CpuConfig.Max >= 0 {
				hasRetry = true
				startCount++
			} else {
				hasRetry = false
				startCount = 0
			}

			// Send mail.
			if hasRetry && startCount == count {
				content := fmt.Sprintf("Alarm: Cpu used  %s%s beycond the threshold! \n from %v ", format.Float64(cpus.CurrentUsage), "%", cf.Addr)
				cpus.CpuMax = 0
				header := "CpuUsed exceed!"
				timestamp := time.Now()
				alert.AlertConvergence(cf.CpuConfig.EmailArray, timestamp, content, header)
				// Begin the next cycle.
				hasRetry = false
				startCount = 0
			}
		}
	}
	return nil
}
