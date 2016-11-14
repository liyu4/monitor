package mem

import (
	"testing"

	"monitor/parse"
)

func TestGetMem(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")
	mem := newMem()

	err := mem.monitorMem()

	if err != nil {
		t.Fatal("Failed get memory monitor data! error: %v", err)
	}
}
