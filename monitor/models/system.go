package models

import (
	"encoding/json"
	"strings"
	"time"

	"monitor/alert"
	"monitor/base"
	"monitor/cpu"
	"monitor/dir"
	"monitor/disk"
	"monitor/file"
	"monitor/format"
	"monitor/mem"
	"monitor/network"
	"monitor/parse"
	"monitor/process"
	"monitor/user"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kevinchen/logx"
	"github.com/kevinchen/stringx"
	"github.com/mygojson"
	"github.com/robfig/cron"
)

// Daily report.
type System struct {
	Id                    int
	CpuCurrent            float64
	CpuMax                float64
	CpuMin                float64
	CpuAver               float64 `json:"omit"`
	DirectoryInfo         string  `orm:"type(text)"`
	FileInfo              string  `orm:"type(text)"`
	DiskInfo              string  `orm:"type(text)"`
	MemTotal              uint64
	MemFree               uint64
	MemCurrentUsedPercent float64
	MemMaxUsedPercent     float64
	MemMinUsedPercent     float64
	MemAverUsedPercent    float64
	NetworkInfos          string `orm:"type(text)"`
	ProcessInfo           string `orm:"type(text)"`
	Userinfos             string `orm:"type(text)"`
	ModuleNames           string
	Inserttime            int64
	Date                  time.Time
}

func init() {
	orm.RegisterModel(new(System))
	cf = parse.NewCf()
}

var cf *parse.Config

var data = make(chan System)

type Sizes struct {
	Sizes []Size
}

type Size struct {
	Name string
	Size string
}

type Disk struct {
	DiskTotal   string
	DiskUsed    string
	DiskFree    string
	DiskPercent string
	Path        string
}

type AllDisk struct {
	Ad []Disk
}

type FileInfo struct {
	Name string
	Size string
}

type FileInfos struct {
	Fl []FileInfo `json:"fileinfos"`
}

type Network struct {
	Name       string
	NetTotalTx string
	NetTotalRx string
}

type Networks struct {
	Nets []Network
}

func insert() {
	o := orm.NewOrm()
	// getCPU := cpu.NewCpu()
	useageOfCPU := cpu.Result()

	// Pointer does not save data
	// Assign default values (-1)
	if useageOfCPU.CpuMax == 0 {
		useageOfCPU.CpuAver = -1
		useageOfCPU.CpuMax = -1
		useageOfCPU.CpuMin = -1
		useageOfCPU.CurrentUsage = -1
	}

	sizeOfDir, err := dir.GetDirSize()
	if err != nil {
		logx.FError("get directory size error %v", err)
	}

	if sizeOfDir.Dirs == nil {
		sizeOfDir.Dirs = []dir.DirInfo{dir.DirInfo{Name: "nodata", Size: -1}}
	}

	subsize := Sizes{}

	for _, v := range sizeOfDir.Dirs {
		subsize.Sizes = append(subsize.Sizes, Size{
			Name: v.Name,
			Size: format.TranslateDir(float64(v.Size)),
		})
	}

	directoyInfo, err := json.Marshal(subsize)

	if err != nil {
		logx.FError("%v", err)
	}

	diskInfo, err := disk.GetDiskUsage()
	if err != nil {
		logx.FError("get disk error %v", err)
	}

	subdisk := AllDisk{}
	for _, v := range diskInfo.Ad {
		if v.DiskTotal == 0 {
			continue
		}
		subdisk.Ad = append(subdisk.Ad, Disk{
			DiskTotal:   format.Translate(float64(v.DiskTotal)),
			DiskUsed:    format.Translate(float64(v.DiskUsed)),
			DiskFree:    format.Translate(float64(v.DiskFree)),
			DiskPercent: format.Float64(v.DiskPercent) + "%",
			Path:        v.Path,
		})
	}

	// Pointer does not save data
	// Assign default values (0)
	if diskInfo.Ad == nil {
		diskInfo.Ad = []disk.Disk{disk.Disk{
			DiskFree:    0,
			DiskPercent: 0,
			DiskTotal:   0,
			DiskUsed:    0,
			Path:        "",
		}}
	}

	di, err := json.Marshal(subdisk)

	if err != nil {
		logx.FError("%v", err)
	}

	getFile, err := file.GetFileSize()

	// Pointer does not save data
	// Assign default values (0)
	if getFile.Fl == nil {
		getFile.Fl = []file.FileInfo{file.FileInfo{
			Name: "",
			Size: 0,
		}}
	}

	if err != nil {
		logx.FError("get file size error %v: ", err)
	}

	subfile := FileInfos{}

	for _, v := range getFile.Fl {
		subfile.Fl = append(subfile.Fl, FileInfo{
			Name: v.Name,
			Size: format.Translate(float64(v.Size)),
		})
	}

	fileInfo, err := json.Marshal(subfile)

	if err != nil {
		logx.FError("%v", err)
	}

	// If pointer does not save data
	// Assign default values (-1)
	getMemory := mem.Result()

	if getMemory.MemMinUsedPercent == 0 {
		getMemory.MemCurrentUsedPercent = -1
		getMemory.MemFree = 0
		getMemory.MemAverUsedPercent = -1
		getMemory.MemMaxUsedPercent = -1
		getMemory.MemMinUsedPercent = -1
		getMemory.MemTotal = 0
	}

	getNetwork, err := network.GetNetwork()

	// Pointer does not save data
	// Assign default values (0)
	if getNetwork.Nets == nil {
		getNetwork.Nets = []network.Network{network.Network{
			Name:       "",
			NetTotalTx: 0,
			NetTotalRx: 0,
			TxSec:      0,
			RxSec:      0,
		}}
	}

	subnet := Networks{}

	for _, v := range getNetwork.Nets {
		subnet.Nets = append(subnet.Nets, Network{
			Name:       v.Name,
			NetTotalTx: format.Translate(float64(v.NetTotalTx)),
			NetTotalRx: format.Translate(float64(v.NetTotalRx)),
		})
	}

	if err != nil {
		logx.FError("get network information error %v: ", err)
	}

	networkInfo, err := json.Marshal(subnet)

	if err != nil {
		logx.FError("%v", err)
	}

	getProcess, err := process.GetMultiPid()

	// Pointer does not save data
	// Assign default values (default process)
	if getProcess.AllProcess == nil {
		getProcess.AllProcess = []process.Single{process.Single{
			Name: "",
			Pid:  -1,
		}}
	}

	if err != nil {
		logx.FError("get process information error %v: ", err)
	}

	processInfo, err := json.Marshal(getProcess)

	if err != nil {
		logx.FError("%v", err)
	}

	getUser, err := user.GetUserStat()

	// Pointer does not save data
	// Assign default values (default user)

	if getUser.AllUser == nil {
		getUser.AllUser = []user.UserStat{user.UserStat{
			User: "",
		}}
	}

	if err != nil {
		logx.FError("get user infomation error %v: ", err)
	}

	userInfo, err := json.Marshal(getUser)

	if err != nil {
		logx.FError("%v", err)
	}

	system := System{
		CpuCurrent:            useageOfCPU.CurrentUsage,
		CpuMax:                useageOfCPU.CpuMax,
		CpuMin:                useageOfCPU.CpuMin,
		CpuAver:               useageOfCPU.CpuAver,
		DirectoryInfo:         string(directoyInfo),
		DiskInfo:              string(di),
		FileInfo:              string(fileInfo),
		MemTotal:              getMemory.MemTotal,
		MemFree:               getMemory.MemFree,
		MemCurrentUsedPercent: getMemory.MemCurrentUsedPercent,
		MemMaxUsedPercent:     getMemory.MemMaxUsedPercent,
		MemMinUsedPercent:     getMemory.MemMinUsedPercent,
		MemAverUsedPercent:    getMemory.MemAverUsedPercent,
		NetworkInfos:          string(networkInfo),
		ProcessInfo:           string(processInfo),
		Userinfos:             string(userInfo),
		ModuleNames:           strings.Join(cf.CommonConfig.ModuleName, "-"),
		Date:                  time.Now(),
	}
	data <- system
	o.Insert(&system)
	useageOfCPU.Clear()
	getMemory.Clear()
}

func Task() {
	logx.FInfo("%v", "statr crontab task!")
	crontab := cf.SpecConfig.Crontab // Every day nine thirty.

	c := cron.New()
	err := c.AddFunc(crontab, insert)
	if err != nil {
		logx.FError("%v", err)
	}

	c.AddFunc(crontab, report)
	if err != nil {
		logx.FError("%v", err)
	}
	c.Start()
}

// Format beautify
func report() {
	var cf = parse.NewCf()
	var basic = base.NewBase()

	if d, ok := <-data; ok {

		var cpustatus string
		currentUsage := format.Float64(d.CpuCurrent) + "%"
		maxUsage := format.Float64(d.CpuMax) + "%"
		minUsage := format.Float64(d.CpuMin) + "%"
		averUsage := format.Float64(d.CpuAver) + "%"

		if cf.CpuConfig == nil {
			logx.FError("%v", "report refer to cf  module but  cpuconfig section is empty!")
			return
		}

		if d.CpuMax >= cf.CpuConfig.Max {
			cpustatus = `<td bgcolor="red">异常</td>`
		} else {
			cpustatus = `<td bgcolor="green">正常</td>`
		}

		var memStatus string
		total := format.Translate(float64(d.MemTotal))
		free := format.Translate(float64(d.MemFree))
		memCurrentUsedPercent := format.Float64(d.MemCurrentUsedPercent) + "%"
		memMaxUsedPercent := format.Float64(d.MemMaxUsedPercent) + "%"
		memMinUsedPercent := format.Float64(d.MemMinUsedPercent) + "%"
		memAverUsedPercent := format.Float64(d.MemAverUsedPercent) + "%"

		if cf.MemoryConfig == nil {
			logx.FError("%v", "report refer to cf  module but  memconfig section is empty!")
			return
		}

		if d.MemMaxUsedPercent >= cf.MemoryConfig.Max {
			memStatus = `<td bgcolor="red">异常</td>`
		} else {
			memStatus = `<td bgcolor="green">正常</td>`
		}

		// directory
		if cf.DirectoryConfig == nil {
			logx.FError("%v", "report refer to cf  module but  directoryconfig section is empty!")
			return
		}
		var dirSizeHtml string
		{
			dir := mygojson.Json(d.DirectoryInfo)

			dirArray, err := dir.Get("Sizes").Array()

			if err != nil {
				logx.FError("Json umarshal directoryinfo failed%v", err)
			}

			for _, v := range dirArray {
				dirMap := v.(map[string]interface{})

				if basic.TranslateToK(dirMap["Size"].(string)) >= cf.DirectoryConfig.Max {
					dirSizeHtml += `<tr><td>` + dirMap["Name"].(string) +
						`</td><td>` + dirMap["Size"].(string) +
						`</td><td bgcolor="red">异常</td</tr>`
				} else {
					dirSizeHtml += `<tr><td>` + dirMap["Name"].(string) +
						`</td><td>` + dirMap["Size"].(string) +
						`</td><td bgcolor="green">正常</td</tr>`
				}
			}
		}

		// disk
		if cf.DiskConfig == nil {
			logx.FError("%v", "report refer to cf  module but  diskconfig section is empty!")
			return
		}
		var diskSizeHtml string
		{
			disk := mygojson.Json(d.DiskInfo)

			diskArray, err := disk.Get("Ad").Array()

			if err != nil {
				logx.FError("Json umarshal diskinfo failed%v", err)
			}

			for _, v := range diskArray {
				diskMap := v.(map[string]interface{})

				if basic.TranslateToK(diskMap["DiskTotal"].(string)) <= cf.DiskConfig.Condition {
					continue
				}

				if basic.TranslateToK(diskMap["DiskFree"].(string)) <= cf.DiskConfig.Free {
					diskSizeHtml += `<tr><td>` + diskMap["DiskTotal"].(string) +
						`</td><td>` + diskMap["DiskUsed"].(string) +
						`</td><td>` + diskMap["DiskFree"].(string) +
						`</td><td>` + diskMap["DiskPercent"].(string) +
						`</td><td>` + diskMap["Path"].(string) +
						`</td></td><td bgcolor="red">异常</td</tr>`
				} else {
					diskSizeHtml += `<tr><td>` + diskMap["DiskTotal"].(string) +
						`</td><td>` + diskMap["DiskUsed"].(string) +
						`</td><td>` + diskMap["DiskFree"].(string) +
						`</td><td>` + diskMap["DiskPercent"].(string) +
						`</td><td>` + diskMap["Path"].(string) +
						`</td></td><td bgcolor="red">正常</td></tr>`
				}
			}
		}

		// file
		if cf.FileConfig == nil {
			logx.FError("%v", "report refer to cf  module but  fileconfig section is empty!")
			return
		}
		var fileSizeHtml string
		{
			file := mygojson.Json(d.FileInfo)

			fileArray, err := file.Get("fileinfos").Array()

			if err != nil {
				logx.FError("Json umarshal fileinfo failed%v", err)
			}

			for _, v := range fileArray {
				fileMap := v.(map[string]interface{})
				if basic.TranslateToK(fileMap["Size"].(string)) >= cf.FileConfig.Max {
					fileSizeHtml += `<tr><td>` + fileMap["Name"].(string) +
						`</td><td>` + fileMap["Size"].(string) +
						`</td></td><td bgcolor="red">异常</td</tr>`
				} else {
					fileSizeHtml += `<tr><td>` + fileMap["Name"].(string) +
						`</td><td>` + fileMap["Size"].(string) +
						`</td></td><td bgcolor="green">正常</td></tr>`
				}
			}
		}

		// network
		var netHtml string
		{
			net := mygojson.Json(d.NetworkInfos)
			netArray, err := net.Get("Nets").Array()
			if err != nil {
				logx.FError("Json umarshal network failed%v", err)
			}
			for _, v := range netArray {
				netMap := v.(map[string]interface{})
				netHtml += `<tr><td>` + netMap["NetTotalTx"].(string) +
					`</td><td>` + netMap["NetTotalRx"].(string) +
					`</td></td><td bgcolor="green">正常</td></tr>`
			}
		}

		// process
		var processHtml string
		{
			pc := mygojson.Json(d.ProcessInfo)

			pcArray, err := pc.Get("AllProcess").Array()

			if err != nil {
				logx.FError("Json umarshal processinfo failed%v", err)
			}

			for _, v := range pcArray {
				pcMap := v.(map[string]interface{})

				if stringx.MustString2(pcMap["Pid"]) == "0" {
					processHtml += `<tr><td>` + pcMap["Name"].(string) +
						`</td><td>` + stringx.MustString2(pcMap["Pid"]) +
						`</td></td><td bgcolor="red">异常</td</tr>`
				} else {
					processHtml += `<tr><td>` + pcMap["Name"].(string) +
						`</td><td>` + stringx.MustString2(pcMap["Pid"]) +
						`</td></td><td bgcolor="green">正常</td</tr>`
				}
			}
		}

		// login
		var loginHtml string
		var exist bool
		{
			login := mygojson.Json(d.Userinfos)

			loginArray, err := login.Get("userinfos").Array()

			if err != nil {
				logx.FError("Json umarshal login failed%v", err)
			}

			for _, v := range loginArray {
				exist = false
				loginMap := v.(map[string]interface{})
				for _, vv := range cf.UserConfig.Userlist {
					if loginMap["user"].(string) == vv {
						exist = true
					}
				}

				if exist {
					loginHtml += `<tr><td>` + loginMap["user"].(string) +
						`</td><td>` + loginMap["terminal"].(string) +
						`</td><td>` + loginMap["host"].(string) +
						`</td><td>` + stringx.MustString2(loginMap["started"]) +
						`</td></td><td bgcolor="green">正常</td</tr>`
				} else {
					loginHtml += `<tr><td>` + loginMap["user"].(string) +
						`</td><td>` + loginMap["terminal"].(string) +
						`</td><td>` + loginMap["host"].(string) +
						`</td><td>` + stringx.MustString(loginMap["started"]) +
						`</td></td><td bgcolor="red">异常</td</tr>`
				}
			}
		}

		body := `
		         <html>
		         <body>
		         <table border="1">
		         <caption>Cpu</caption>
		         <tr>
		         <th>currentUsage</th>
		         <th>maxUsage</th>
		         <th>minUsage</th>
		         <th>averUsage</th>
				 <th>Status</th>		       
		         </tr>
		         <tr>
		         <td>` + currentUsage + `</td>
		         <td>` + maxUsage + `</td>
		         <td>` + minUsage + `</td>
		         <td>` + averUsage + `</td>
		         ` + cpustatus + `
		         </tr>
		         </table>

		         </br>
		         <table border="1">
		         <caption>Mermory</caption>
		         <tr>
		         <th>Total</th>
		         <th>Free</th>
		         <th>memCurrentUsedPercent</th>
		         <th>memMaxUsedPercent</th>
		         <th>memMinUsedPercent</th>
		         <th>memAverUsedPercent</th>
		         <th>status</th>		       
		         </tr>	
		         <tr>
		         <td>` + total + `</td>
		         <td>` + free + `</td>
		         <td>` + memCurrentUsedPercent + `</td>
		         <td>` + memMaxUsedPercent + `</td>
		         <td>` + memMinUsedPercent + `</td>
		         <td>` + memAverUsedPercent + `</td>
		         ` + memStatus + `
		         </tr>
		         </table>

		         </br>
				 <table border="1">
		         <caption>Directory</caption>                  
		         <tr>
		         <th>Name</th>
		         <th>Size</th>
		         <th>Status</th>
		         </tr>
		         ` + dirSizeHtml + `
                 </table>
				
                 </br>
                 <table border="1">
				 <caption>Disk</caption>
		         <tr>
		         <th>diskTotal</th>
		         <th>diskUsed</th>
		         <th>DiskFree</th>
		         <th>DiskPercent</th>
		         <th>Path</th>
		         <th>Status</th>
		         </tr>
		         <tr>
		         ` + diskSizeHtml + `	                  
		         </tr>
                 </table>

		         </br>
				 <table border="1">
		         <caption>File</caption>                  
		         <tr>
		         <th>Name</th>
		         <th>Size</th>
		         <th>Status</th>
		         </tr>
		         ` + fileSizeHtml + `
                 </table>

                 </br>
				 <table border="1">
		         <caption>Network</caption>                  
		         <tr>
		         <th>NetTotalTx</th>
		         <th>NetTotalRx</th>
		         <th>Status</th>
		         </tr>
		         ` + netHtml + `
                 </table>                 

                 </br>
				 <table border="1">
		         <caption>Process</caption>                  
		         <tr>
		         <th>Name</th>
		         <th>Pid</th>
		         <th>Status</th>
		         </tr>
		         ` + processHtml + `
                 </table> 

                 </br>
		         <table border="1">
		         <caption>Login</caption>
		         <tr>
		         <th>user</th>
		         <th>terminal</th>
		         <th>host</th>
		         <th>started</th>
		         <th>Status</th>
		         </tr>	
		         ` + loginHtml + `
		         </table>            
		         </body>

		         </br>
		         <table border="1">
		         <caption>Application</caption>
		         <tr>
		         <th>name</th>
		         <th>Status</th>
		         </tr>	
		         <tr>
		         <td>
		         ` + d.ModuleNames + `
		         </td>
		         <td bgcolor="green">
		         正常
		         </td>
		         </tr>
		         </table>            

		         </html>
   	             `

		header := "Report!"
		alert.OnlySend(cf.ReportConfig.EmailArray, body, header)
	}
}
