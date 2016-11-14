package process

import (
	"fmt"
	"time"

	"monitor/alert"
	"monitor/common"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/kevinchen/numberx"
	"github.com/shirou/gopsutil/process"
)

// Need monitor process name.

type Single struct {
	Name string
	Pid  int
}

type Process struct {
	AllProcess []Single
}

func newProcessName() *Process {
	return &Process{}
}

func GetMultiPid() (*Process, error) {
	pro := newProcessName()
	return pro.getMultiPid()
}

//If need get all pids.
func getAllPids() ([]int32, error) {
	allPid, err := process.Pids()
	if err != nil {
		return nil, err
	}
	return allPid, nil
}

func (p *Process) getMultiPid() (*Process, error) {
	var cf = parse.NewCf()

	for _, name := range cf.LocalPidListConfig.NameList {
		pid, err := common.GetPid(name)

		if err != nil {
			logx.FError("getMultiPid get procee pid info error: %v", err)
			return &Process{}, err
		}

		p.AllProcess = append(p.AllProcess, Single{
			Name: name,
			Pid:  numberx.MustInt(pid, 0),
		})
	}
	return p, nil
}

func MonitorLocalPid() {

	go func() {
		process := newProcessName()
		logx.FInfo("%v", "Process module started!")
		if err := process.monitorLocalPid(); err != nil {
			logx.FCritical("MonitorLocalPid error: %v", err)
		}
	}()

}

func (p *Process) monitorLocalPid() error {
	var cf = parse.NewCf()

	if cf.LocalPidListConfig == nil {
		logx.FError("%v", "LocalPidListConfig is nil!")
		return fmt.Errorf("LocalPidListConfig is nil!")
	}

	silence := cf.LocalPidListConfig.Duration

	if silence < 500 {
		silence = 500
	}

	t1 := time.NewTicker(time.Millisecond * time.Duration(silence))
	count := (cf.LocalPidListConfig.Lasttime * (1000)) / cf.LocalPidListConfig.Duration
	collection := make(map[string]int)
	p.reset()
	for {
		select {
		case <-t1.C:
			process, err := p.getMultiPid()
			if err != nil {
				continue
			}

			// Is zero for does not exist.
			for _, v := range process.AllProcess {
				logx.FDebug("Process info  Application_name: %v Pid: %v", v.Name, v.Pid)
				if v.Pid == 0 {
					collection[v.Name] += 1
				} else {
					collection[v.Name] = 0
				}

				for name, num := range collection {
					if num == count {
						content := fmt.Sprintf("Alarm: process %s can not be found! \nfrom: %v", name, cf.Addr)
						timestamp := time.Now()
						header := "Application " + name + " can not be found!"
						alert.AlertConvergence(cf.LocalPidListConfig.EmailArray, timestamp, content, header)
						// Begin the next cycle.
						collection[name] = 0
					}
				}
			}
			p.reset()
		}
	}
	return nil
}

func (p *Process) reset() {
	p.AllProcess = p.AllProcess[:0:0]
}
