package file

import (
	"fmt"
	"time"

	"monitor/alert"
	"monitor/base"
	"monitor/format"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/kevinchen/numberx"
)

/*
   At the end of the day ,when it comes down to it, all we
   really want is to be close to somebody.
*/

//if need monitor multi file.
type FileInfo struct {
	Name string
	Size int64
}

type FileInfos struct {
	Fl []FileInfo `json:"fileinfos"`
}

func newFileInfos() *FileInfos {
	return &FileInfos{}
}

func GetFileSize() (*FileInfos, error) {
	file := newFileInfos()
	return file.getFileSize()
}

func (f *FileInfos) getFileSize() (*FileInfos, error) {
	var cf = parse.NewCf()

	if cf.FileConfig == nil {
		logx.FError("%v", "FileConfig is nil!")
		return &FileInfos{}, fmt.Errorf("FileConfig is nil!")
	}

	bs := base.NewBase()
	for _, path := range cf.FileConfig.Filelist {

		raw, err := bs.ProcessFile(path)

		if err != nil {
			logx.FError("getFilesize's processfile error  %v", err)
			return &FileInfos{}, err
		}

		singleFileinfo := FileInfo{
			Name: path,
			Size: numberx.MustInt64(raw, 0),
		}

		f.Fl = append(f.Fl, singleFileinfo)
	}
	return f, nil
}

func MonitorFile() {
	go func() {
		file := newFileInfos()
		logx.FInfo("%v", "File module started!")
		err := file.monitorFile()
		if err != nil {
			logx.FCritical("%v", err)
		}
	}()
}

func (f *FileInfos) monitorFile() error {
	var cf = parse.NewCf()

	if cf.FileConfig == nil {
		logx.FError("%v", "FileConfig is nil!")
		return fmt.Errorf("FileConfig is nil!")
	}

	f.reset()
	t1 := time.NewTicker(time.Second * time.Duration(cf.FileConfig.Lasttime))
	for {
		select {
		case <-t1.C:
			files, err := f.getFileSize()

			if err != nil {
				continue
			}

			for _, v := range files.Fl {
				logx.FDebug("File info filename: %v filesize: %v ", v.Name, format.Translate(float64(v.Size)))
				if v.Size >= cf.FileConfig.Max {
					content := fmt.Sprintf("Alarm: monitor file %s used  %s beycond the threshold! \nfrom: %v", v.Name, format.Translate(float64(v.Size)), cf.Addr)
					header := v.Name + " used has exceeded the maximum!"
					timestamp := time.Now()
					alert.AlertConvergence(cf.FileConfig.EmailArray, timestamp, content, header)
				}
			}
			f.reset()
		}
	}
	return nil
}

func (f *FileInfos) reset() {
	f.Fl = f.Fl[:0:0]
}
