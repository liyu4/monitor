package mem

import (
	"fmt"
	"sync"
	"time"

	"monitor/alert"
	"monitor/format"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/shirou/gopsutil/mem"
)

type Mem struct {
	//Fixed-size.
	MemTotal uint64
	MemFree  uint64

	MemCurrentUsedPercent float64
	MemMaxUsedPercent     float64
	MemMinUsedPercent     float64
	MemAverUsedPercent    float64
}

var (
	record *Mem
	lock   = sync.RWMutex{}
)

// Factory model.
func newMem() *Mem {
	return &Mem{}
}

func (m *Mem) getMemUsage() error {
	memory, err := mem.VirtualMemory()
	if err != nil {
		logx.FError("getMemUsage to get virtural memmory error: %v", err)
		return err
	}
	m.MemTotal = memory.Total
	m.MemFree = memory.Free

	if m.MemCurrentUsedPercent == 0 && m.MemMaxUsedPercent == 0 &&
		m.MemMinUsedPercent == 0 && m.MemAverUsedPercent == 0 {
		m.MemCurrentUsedPercent = memory.UsedPercent
		m.MemMaxUsedPercent = memory.UsedPercent
		m.MemMinUsedPercent = memory.UsedPercent
		m.MemAverUsedPercent = memory.UsedPercent
	}

	if memory.UsedPercent > m.MemMaxUsedPercent {
		m.MemMaxUsedPercent = memory.UsedPercent
		m.MemCurrentUsedPercent = memory.UsedPercent
	} else {
		if m.MemMinUsedPercent > memory.UsedPercent {
			m.MemMinUsedPercent = memory.UsedPercent
		}
		m.MemCurrentUsedPercent = memory.UsedPercent
	}

	if m.MemAverUsedPercent != memory.UsedPercent {
		m.MemAverUsedPercent = (m.MemAverUsedPercent + memory.UsedPercent) / 2
	}

	return nil
}

func MonitorMem() {
	go func() {
		mem := newMem()
		logx.FInfo("%v", "Memory module started!")
		err := mem.monitorMem()
		if err != nil {
			logx.FCritical("MonitorMem error: %v", err)
		}
	}()
}

func Result() *Mem {
	return record
}

func (m *Mem) Clear() {
	m.MemTotal = 0
	m.MemFree = 0

	m.MemCurrentUsedPercent = 0
	m.MemMaxUsedPercent = 0
	m.MemMinUsedPercent = 0
	m.MemAverUsedPercent = 0

}

//Monitor local loadbalance.
func (m *Mem) monitorMem() error {
	var cf = parse.NewCf()

	if cf.MemoryConfig == nil {
		logx.FError("%v", "MemoryConfig is nil!")
		return fmt.Errorf("MemoryConfig is nil!")
	}

	t1 := time.NewTicker(time.Millisecond * time.Duration(cf.MemoryConfig.Duration))
	hasRetry := false
	count := (cf.MemoryConfig.Lasttime * (1000)) / cf.MemoryConfig.Duration
	startCount := 0
	for {
		select {
		case <-t1.C:
			err := m.getMemUsage()
			if err != nil {
				continue
			}

			logx.FDebug("Memory info Total: %v Free: %v MaxPercent: %v%s MinPercent: %v%s Average: %v%s Current: %v%s",
				format.Translate(float64(m.MemTotal)), format.Translate(float64(m.MemFree)), format.Float64(m.MemMaxUsedPercent), "%", format.Float64(m.MemMinUsedPercent), "%",
				format.Float64(m.MemAverUsedPercent), "%", format.Float64(m.MemCurrentUsedPercent), "%",
			)

			lock.Lock()
			record = m
			lock.Unlock()

			// For test.
			if m.MemCurrentUsedPercent-cf.MemoryConfig.Max >= 0 {
				hasRetry = true
				startCount++
			} else {
				hasRetry = false
				startCount = 0
			}

			// Send mail.
			if hasRetry && startCount == count {
				content := fmt.Sprintf("Alarm: Memory used  %s%s beycond the threshold! \nform: %v", format.Float64(m.MemCurrentUsedPercent), "%", cf.Addr)
				timestamp := time.Now()
				header := "Memory exceed!"
				alert.AlertConvergence(cf.MemoryConfig.EmailArray, timestamp, content, header)
				// Begin the next cycle.
				hasRetry = false
				startCount = 0
			}
		}
	}
	return nil
}
