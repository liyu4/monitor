package base

/* 终日昏昏醉梦间
   忽闻春尽强登山
   因过竹院逢僧话
   偷的浮生半日闲
*/

import (
	"context"
	"errors"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/kevinchen/filepathx"
	"github.com/kevinchen/numberx"

	config "github.com/Unknwon/goconfig"
)

type Base struct {
}

func NewBase() *Base {
	return &Base{}
}

func (b *Base) CheckString(source string) error {
	if source == "" {
		return errors.New("Parse string type error,check the key corresponding to value! ")
	}
	return nil
}

func (b *Base) CheckInt(source int) error {
	if source == 0 {
		return errors.New("Parse int type error,check the key corresponding to value! ")
	}
	return nil
}

func (b *Base) CheckInt64(source int64) error {
	if source == 0 {
		return errors.New("Parse int64 type error,check the key corresponding to value! ")
	}
	return nil
}

func (b *Base) CheckFloat64(source float64) error {
	if source == 0 {
		return errors.New("Parse float64 type error,check the key corresponding to value! ")
	}
	return nil
}

func (b *Base) CheckArray(source []string) error {
	if len(source) == 0 {
		return errors.New("Parse Array type error,check the key corresponding to value! ")
	}
	return nil
}

func (b *Base) CheckFile(path string) error {
	if filepathx.IsFile(path) {
		return nil
	}
	return errors.New("no such file! ")
}

func (b *Base) CheckDir(dir string) error {
	if filepathx.IsDir(dir) {
		return nil
	}
	return errors.New("no such directory! ")
}

func (b *Base) ProcessFile(file string) (string, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", `ls -al `+file+`|awk '{print $5}'`)

	data, err := cmd.StdoutPipe()

	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	raw, err := ioutil.ReadAll(data)

	if err != nil {
		return "", err
	}

	cmd.Wait()
	cancel()

	select {
	case <-ctx.Done():
	}

	return string(raw), nil
}

func (b *Base) ProcessDir(dir string) (string, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	cmd := exec.CommandContext(ctx, "du", "-sh", dir)

	data, err := cmd.StdoutPipe()

	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	raw, err := ioutil.ReadAll(data)

	if err != nil {
		return "", err
	}
	cmd.Wait()
	cancel()
	select {
	case <-ctx.Done():
	}

	return string(raw), err
}

func (b *Base) TranslateToK(data string) int64 {
	data = strings.TrimSpace(data)
	data = strings.ToUpper(data)

	if strings.Contains(data, "K") {
		temp := strings.TrimSuffix(data, "K")
		if strings.Contains(temp, ".") {
			result := strings.Split(temp, ".")
			return numberx.MustInt64(result[0], 0)
		} else {
			return numberx.MustInt64(temp, 0)
		}
	}

	if strings.Contains(data, "M") {
		temp := strings.TrimSuffix(data, "M")
		if strings.Contains(temp, ".") {
			result := strings.Split(temp, ".")
			return numberx.MustInt64(result[0], 0)*1024 + numberx.MustInt64(result[1], 0)*1024/(divider(result[1]))
		}
		return numberx.MustInt64(temp, 0) * 1024
	}

	if strings.Contains(data, "G") {
		temp := strings.TrimSuffix(data, "G")
		if strings.Contains(temp, ".") {
			result := strings.Split(temp, ".")
			return numberx.MustInt64(result[0], 0)*1024*1024 + numberx.MustInt64(result[1], 0)*1048576/divider(result[1])
		} else {
			return numberx.MustInt64(temp, 0) * 1024 * 1024
		}
	}
	return 0
}

func NewConfig() (*config.ConfigFile, error) {
	cfg, err := config.LoadConfigFile("/Users/admin/svn/src/monitor/conf/app.conf")
	return cfg, err
}

func divider(result string) int64 {
	var divider int64 = 1
	for i := 0; i < len(result); i++ {
		divider *= 10
	}
	return divider
}
