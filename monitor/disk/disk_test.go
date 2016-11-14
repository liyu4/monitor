package disk

import (
	"testing"

	"monitor/parse"
)

func TestGetDisk(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")

	d := newAllDisk()

	err := d.monitorAllDisk()

	if err != nil {
		t.Fatal("Failed get disk monitor data! error: %v", err)
	}
}
