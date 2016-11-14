package dir

import (
	"testing"

	"monitor/parse"
)

func TestGetDirSize(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")
	d := newDirInfos()

	err := d.monitorDir()
	if err != nil {
		t.Fatal("Failed get directory monitor data! error: %v", err)
	}
}
