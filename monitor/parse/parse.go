package parse

import (
	"net"
	"sync"

	"monitor/base"

	config "github.com/Unknwon/goconfig"
	"github.com/kevinchen/filepathx"
	"github.com/kevinchen/logx"
)

type Common struct {
	Duration   int
	ModuleName []string
}

type Cpu struct {
	Max        float64
	Duration   int
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type Memory struct {
	Max        float64
	Duration   int
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type Disk struct {
	Condition  int64
	Free       int64
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type Net struct {
	MaxRx      int64
	MaxTx      int64
	Duration   int
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type LocalPidList struct {
	NameList   []string
	Duration   int
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type Directory struct {
	DirList    []string
	Max        int64
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type File struct {
	Filelist   []string
	Max        int64
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type User struct {
	Userlist   []string
	Lasttime   int
	Group      string
	EmailArray []string
	Sync       bool
}

type Report struct {
	Group      string
	EmailArray []string
	Sync       bool
}

type Socket struct {
	Addr       string
	Port       string
	Group      string
	EmailArray []string
	Timeout    int64
	Sync       bool
}

type Http struct {
	Group      string
	EmailArray []string
	Timeout    int64
	Sync       bool
}

type Spec struct {
	Crontab string
}

type Module struct {
	Name       string
	Group      string
	EmailArray []string
}

type Modules struct {
	Ms []Module
}

type Config struct {
	SpecConfig         *Spec
	HttpConfig         *Http
	SocketConfig       *Socket
	CommonConfig       *Common
	CpuConfig          *Cpu
	MemoryConfig       *Memory
	DiskConfig         *Disk
	NetConfig          *Net
	LocalPidListConfig *LocalPidList
	DirectoryConfig    *Directory
	FileConfig         *File
	UserConfig         *User
	ReportConfig       *Report
	ModuleConfig       Modules
	lock               sync.RWMutex
	Separate           string
	Config             *config.ConfigFile
	Addr               string
}

var (
	bs = base.NewBase()
	cf *Config
)

func (c *Config) Start(fileconfig string) {
	var err error
	execDirAbsPath := filepathx.AppRootDir()
	if fileconfig == "" {
		c.Config, err = config.LoadConfigFile(execDirAbsPath + "/conf/app.conf")
	} else {
		c.Config, err = config.LoadConfigFile(fileconfig)
	}

	if err != nil {
		logx.FError("load config file failed: %v", err)
		return
	}

	if err := c.getCpu(); err != nil {
		logx.FError("get cpu config error:%v", err)
	}

	if err := c.getDiretory(); err != nil {
		logx.FError("get directory config error:%v", err)

	}

	if err := c.getDisk(); err != nil {
		logx.FError("get disk config error:%v", err)

	}

	if err := c.getFile(); err != nil {
		logx.FError("get file config error:%v", err)

	}
	if err := c.getMemory(); err != nil {
		logx.FError("get memory config error:%v", err)

	}

	if err := c.getNetwork(); err != nil {
		logx.FError("get network config error:%v", err)

	}

	if err := c.getProcess(); err != nil {
		logx.FError("get process config error:%v", err)

	}

	if err := c.getUser(); err != nil {
		logx.FError("get user config error:%v", err)
	}

	if err := c.getReport(); err != nil {
		logx.FError("get report config error:%v", err)
	}

	if err := c.getCommon(); err != nil {
		logx.FError("get common config error:%v", err)
	}

	if err := c.getSocket(); err != nil {
		logx.FError("get socket config error:%v", err)
	}

	if err := c.getHttp(); err != nil {
		logx.FError("get http config error: %v", err)
	}

	if err := c.getSpec(); err != nil {
		logx.FError("get spec config error: %v", err)
	}

	if err := c.getModule(); err != nil {
		logx.FError("get module config error: %v", err)
	}
	c.getIP()
}

func NewCf() *Config {

	if cf == nil {
		cf = &Config{Separate: "#$#"}
		cf.lock.Lock()
		cf.Start("")
		cf.lock.Unlock()
	}

	return cf
}

func UpdateCf(configfile string) *Config {
	if cf == nil {
		cf = &Config{Separate: "#$#"}
	}
	cf.lock.Lock()
	cf.Start(configfile)
	cf.lock.Unlock()

	return cf
}

func (c *Config) Update() {
	c.lock.Lock()
	c.Start("")
	c.lock.Unlock()
}

func (c *Config) getCpu() error {
	sync := c.Config.MustBool("cpu", "turn_on", true)

	if !sync {
		c.CpuConfig = &Cpu{}
		return nil
	}

	duration := c.Config.MustInt("cpu", "scan_duration")

	if err := bs.CheckInt(duration); err != nil {
		return err
	}

	lasttime := c.Config.MustInt("cpu", "lasttime")

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	maxused := c.Config.MustFloat64("cpu", "max_used")

	if err := bs.CheckFloat64(maxused); err != nil {
		return err

	}

	mailGroupKey := c.Config.MustValue("cpu", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.CpuConfig = &Cpu{
		Max:        maxused,
		Duration:   duration,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getDiretory() error {

	sync := c.Config.MustBool("dir", "turn_on", true)

	if !sync {
		c.DirectoryConfig = &Directory{}
		return nil
	}

	dirArray := c.Config.MustValueArray("dir", "dirlist", cf.Separate)

	if err := bs.CheckArray(dirArray); err != nil {
		return err

	}

	var i int = 0
	var err error
	var newDirArray = make([]string, 0)
	for _, v := range dirArray {
		if err = bs.CheckDir(v); err != nil {
			i++
			logx.FError("parse directory walk %v error: %v", v, err)
			continue
		}

		newDirArray = append(newDirArray, v)
	}

	if i == len(dirArray) {
		return err
	}

	lasttime := c.Config.MustInt("dir", "lasttime")

	if err := bs.CheckInt(lasttime); err != nil {
		return err

	}

	maxsize := c.Config.MustInt64("dir", "maxsize")

	if err := bs.CheckInt64(maxsize); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("dir", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.DirectoryConfig = &Directory{
		DirList:    dirArray,
		Max:        maxsize,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}

	return nil
}

func (c *Config) getDisk() error {
	sync := c.Config.MustBool("disk", "turn_on", true)

	if !sync {
		c.DiskConfig = &Disk{}
		return nil
	}
	condition := c.Config.MustInt64("disk", "condition")

	if err := bs.CheckInt64(condition); err != nil {
		return err
	}

	lasttime := c.Config.MustInt("disk", "lasttime")

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	maxsize := c.Config.MustInt64("disk", "free")

	if err := bs.CheckInt64(maxsize); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("disk", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.DiskConfig = &Disk{
		Condition:  condition,
		Free:       maxsize,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getFile() error {

	sync := c.Config.MustBool("files", "turn_on", true)

	if !sync {
		c.FileConfig = &File{}
		return nil
	}

	files := c.Config.MustValueArray("files", "filelist", cf.Separate)

	if err := bs.CheckArray(files); err != nil {
		return err
	}

	var newfiles = make([]string, 0)

	var i int = 0
	var err error
	for _, v := range files {
		if err = bs.CheckFile(v); err != nil {
			i++
			logx.FError("parse file walk %v error: %v", v, err)
			continue
		}

		newfiles = append(newfiles, v)
	}

	if i == len(files) {
		return err
	}

	lasttime := c.Config.MustInt("files", "lasttime")

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	maxsize := c.Config.MustInt64("files", "maxsize")

	if err := bs.CheckInt64(maxsize); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("files", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.FileConfig = &File{
		Filelist:   newfiles,
		Max:        maxsize,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getMemory() error {
	sync := c.Config.MustBool("mem", "turn_on", true)

	if !sync {
		c.MemoryConfig = &Memory{}
		return nil
	}
	duration := c.Config.MustInt("mem", "scan_duration")

	if err := bs.CheckInt(duration); err != nil {
		return err
	}

	lasttime := c.Config.MustInt("mem", "lasttime")

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	maxused := c.Config.MustFloat64("mem", "max_used")

	if err := bs.CheckFloat64(maxused); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("mem", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.MemoryConfig = &Memory{
		Max:        maxused,
		Duration:   duration,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}

	return nil
}

func (c *Config) getNetwork() error {
	sync := c.Config.MustBool("net", "turn_on", true)

	if !sync {
		c.NetConfig = &Net{}
		return nil
	}
	rx := c.Config.MustInt64("net", "max_bytes_rx_sec", 0)

	if err := bs.CheckInt64(rx); err != nil {
		return err
	}

	tx := c.Config.MustInt64("net", "max_bytes_tx_sec", 0)

	if err := bs.CheckInt64(tx); err != nil {
		return err
	}

	duration := c.Config.MustInt("net", "scan_duration", 0)

	if err := bs.CheckInt(duration); err != nil {
		return err
	}

	lasttime := c.Config.MustInt("net", "lasttime", 0)

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("net", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.NetConfig = &Net{
		MaxRx:      rx,
		MaxTx:      tx,
		Duration:   duration,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getProcess() error {

	sync := c.Config.MustBool("localpidlist", "turn_on", true)

	if !sync {
		c.LocalPidListConfig = &LocalPidList{}
		return nil
	}

	pidArray := c.Config.MustValueArray("localpidlist", "namelist", cf.Separate)

	if err := bs.CheckArray(pidArray); err != nil {
		return err
	}

	duration := c.Config.MustInt("localpidlist", "scan_duration")

	if err := bs.CheckInt(duration); err != nil {
		return err

	}

	lasttime := c.Config.MustInt("localpidlist", "lasttime")

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("localpidlist", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.LocalPidListConfig = &LocalPidList{
		NameList:   pidArray,
		Duration:   duration,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}

	return nil
}

func (c *Config) getUser() error {
	sync := c.Config.MustBool("user", "turn_on", true)

	if !sync {
		c.UserConfig = &User{}
		return nil
	}

	lasttime := c.Config.MustInt("user", "lasttime", 0)

	if err := bs.CheckInt(lasttime); err != nil {
		return err
	}

	mailGroupKey := c.Config.MustValue("user", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	userArray := c.Config.MustValueArray("user", "usernamelist", cf.Separate)

	if err := bs.CheckArray(userArray); err != nil {
		return err
	}

	c.UserConfig = &User{
		Userlist:   userArray,
		Lasttime:   lasttime,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getCommon() error {
	duattion := c.Config.MustInt("common", "scan_duration", 0)

	if err := bs.CheckInt(duattion); err != nil {
		return err
	}

	applications := c.Config.MustValueArray("common", "module_name", c.Separate)

	if err := bs.CheckArray(applications); err != nil {
		return err
	}

	c.CommonConfig = &Common{
		Duration:   duattion,
		ModuleName: applications,
	}
	return nil
}

func (c *Config) getModule() error {
	c.ModuleConfig.Ms = c.ModuleConfig.Ms[:0:0]
	applications := c.Config.MustValueArray("common", "module_name", c.Separate)

	if err := bs.CheckArray(applications); err != nil {
		return err
	}

	for _, v := range applications {
		key := c.Config.MustValue(v, "email", "")

		if err := bs.CheckString(key); err != nil {
			return err
		}

		mailArray := cf.Config.MustValueArray(key, "emails", cf.Separate)

		if err := bs.CheckArray(mailArray); err != nil {
			return err
		}

		c.ModuleConfig.Ms = append(c.ModuleConfig.Ms, Module{
			Name:       v,
			Group:      key,
			EmailArray: mailArray,
		})
	}
	return nil
}

func (c *Config) getReport() error {
	sync := c.Config.MustBool("report", "turn_on", true)

	if !sync {
		c.ReportConfig = &Report{}
		return nil
	}

	mailGroupKey := c.Config.MustValue("report", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	c.ReportConfig = &Report{
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getSocket() error {
	sync := c.Config.MustBool("socket", "turn_on", true)

	if !sync {
		c.SocketConfig = &Socket{}
		return nil
	}
	mailGroupKey := c.Config.MustValue("socket", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	timeout := c.Config.MustInt64("socket", "timeout", 0)

	if err := bs.CheckInt64(timeout); err != nil {
		return err
	}

	addr := c.Config.MustValue("socket", "addr")

	if err := bs.CheckString(addr); err != nil {
		return err
	}

	port := c.Config.MustValue("socket", "port")

	if err := bs.CheckString(port); err != nil {
		return err
	}

	c.SocketConfig = &Socket{
		Addr:       addr,
		Port:       port,
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Timeout:    timeout,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getHttp() error {

	sync := c.Config.MustBool("http", "turn_on", true)

	if !sync {
		c.HttpConfig = &Http{}
		return nil
	}

	mailGroupKey := c.Config.MustValue("http", "email")

	if err := bs.CheckString(mailGroupKey); err != nil {
		return err
	}

	mailArray := c.Config.MustValueArray(mailGroupKey, "emails", cf.Separate)

	if err := bs.CheckArray(mailArray); err != nil {
		return err
	}

	timeout := c.Config.MustInt64("http", "timeout", 0)

	if err := bs.CheckInt64(timeout); err != nil {
		return err
	}

	c.HttpConfig = &Http{
		Group:      mailGroupKey,
		EmailArray: mailArray,
		Timeout:    timeout,
		Sync:       sync,
	}
	return nil
}

func (c *Config) getSpec() error {
	crontab := c.Config.MustValue("spec", "crontab")

	if err := bs.CheckString(crontab); err != nil {
		return err
	}

	c.SpecConfig = &Spec{
		Crontab: crontab,
	}
	return nil
}

func (c *Config) getIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logx.FError("getIp error: %v", err)
		return
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				c.Addr = ipnet.IP.String()
				return
			}
		}
	}
	c.Addr = "Unkonw Host addr!"
}
