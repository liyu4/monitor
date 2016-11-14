package process

import (
	"testing"

	"monitor/parse"
)

func TestGetMultiPid(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")
	process := newProcessName()

	err := process.monitorLocalPid()

	if err != nil {
		t.Fatal("Failed get process monitor data! error: %v", err)
	}
}
