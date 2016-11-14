package main

import (
	"monitor/alert"
	"monitor/controllers"
	"monitor/cpu"
	"monitor/dir"
	"monitor/disk"
	"monitor/file"
	"monitor/mem"
	"monitor/models"
	"monitor/network"
	"monitor/parse"
	"monitor/process"
	_ "monitor/routers"
	"monitor/socket"
	"monitor/user"
	"runtime"

	"github.com/astaxie/beego"
	"github.com/kevinchen/ormx"
)

func main() {
	var cf = parse.NewCf()

	if cf.ReportConfig.Sync {
		ormx.Connect()
		ormx.Syncdb()
		models.Task()
	}

	if cf.CpuConfig != nil {
		if cf.CpuConfig.Sync {
			cpu.MonitorCpu()
		}
	}

	if cf.DirectoryConfig != nil {
		if cf.DirectoryConfig.Sync {
			dir.MonitorDir()
		}
	}

	if cf.DiskConfig != nil {
		if cf.DiskConfig.Sync {

			disk.MonitorAllDisk()
		}
	}

	if cf.FileConfig != nil {
		if cf.FileConfig.Sync {
			file.MonitorFile()
		}

	}

	if cf.MemoryConfig != nil {
		if cf.MemoryConfig.Sync {
			mem.MonitorMem()
		}
	}

	if cf.NetConfig != nil {
		if cf.NetConfig.Sync {
			network.MonitorNetwork()
		}
	}

	if cf.LocalPidListConfig != nil {
		if cf.LocalPidListConfig.Sync {
			process.MonitorLocalPid()
		}
	}

	if cf.UserConfig != nil {
		if cf.UserConfig.Sync {
			user.MonitorUsers()
		}
	}

	if cf.SocketConfig != nil {
		if cf.SocketConfig.Sync {
			socket.Run()
		}
	}

	if cf.HttpConfig != nil {
		if cf.HttpConfig.Sync {
			controllers.Isoff()
		}
	}
	alert.MialQueueTask()
	runtime.GOMAXPROCS(1)
	beego.Run()
}
