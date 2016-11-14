package cpu

import (
	"testing"

	"monitor/parse"
)

func TestGet(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")
	gcu := NewCpu()

	err := gcu.monitorCpu()

	if err != nil {
		t.Fatal("Failed get cpu usage by uderlying! error: %v", err)
	}
}
