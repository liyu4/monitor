package dir

import (
	"fmt"
	"strings"
	"time"

	"monitor/alert"
	"monitor/base"
	"monitor/format"
	"monitor/parse"

	"github.com/kevinchen/logx"
)

type DirInfo struct {
	Name string
	Size int64
}

type DirInfos struct {
	Dirs []DirInfo `json:"directoryinfos"`
}

func newDirInfos() *DirInfos {
	return &DirInfos{}
}

func GetDirSize() (*DirInfos, error) {
	dir := newDirInfos()
	return dir.getDirSize()
}

func (d *DirInfos) getDirSize() (*DirInfos, error) {
	var cf = parse.NewCf()

	if cf.DirectoryConfig == nil {
		logx.FError("%v", "DirectoryConfig is nil!")
		return &DirInfos{}, fmt.Errorf("DirectoryConfig is nil!")
	}

	bs := base.NewBase()
	for _, dir := range cf.DirectoryConfig.DirList {

		dirSize, err := bs.ProcessDir(dir)

		if err != nil {
			logx.FError("getDirSize's processDir error: %v", err)
			return &DirInfos{}, err
		}

		data := strings.Split(dirSize, "/")

		if len(data) == 0 {
			return &DirInfos{}, fmt.Errorf("%v", "invalid data!")
		}

		size := bs.TranslateToK(data[0])

		if size == 0 {
			logx.FError("%v", "getDirSize: Get diectory  is faild ")
			return &DirInfos{}, fmt.Errorf("%v", "getDirSize: Get diectory is faild ")
		}

		dirinfo := DirInfo{
			Name: dir,
			Size: size,
		}

		d.Dirs = append(d.Dirs, dirinfo)
	}
	return d, nil
}

func MonitorDir() {
	go func() {
		dir := newDirInfos()
		logx.FInfo("%v", "Directory module started!")
		err := dir.monitorDir()
		if err != nil {
			logx.FCritical("MonitorDir error: %v", err)
		}
	}()
}

func (d *DirInfos) monitorDir() error {
	var cf = parse.NewCf()

	if cf.DirectoryConfig == nil {
		logx.FError("%v", "DirectoryConfig is nil!")
		return fmt.Errorf("DirectoryConfig is nil!")
	}

	//Clean slice.
	d.reset()

	t1 := time.NewTicker(time.Second * time.Duration(cf.DirectoryConfig.Lasttime))
	for {
		select {
		case <-t1.C:
			dirs, err := d.getDirSize()

			if err != nil {
				continue
			}

			for _, v := range dirs.Dirs {
				logx.FDebug("dirctory infos  name: %v size: %v", v.Name, format.TranslateDir(float64(v.Size)))
				if v.Size >= cf.DirectoryConfig.Max {
					content := fmt.Sprintf("Alarm: monitor's directory %s used  %s beycond the threshold! \nfrom: %v ", v.Name, format.TranslateDir(float64(v.Size)), cf.Addr)
					header := v.Name + " being monitored has exceeded the maximum!"
					timestamp := time.Now()
					alert.AlertConvergence(cf.DirectoryConfig.EmailArray, timestamp, content, header)
				}
			}
			d.reset()
		}
	}
	return nil
}

func (d *DirInfos) reset() {
	d.Dirs = d.Dirs[:0:0]
}
