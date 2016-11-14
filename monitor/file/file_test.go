package file

import (
	"testing"

	"monitor/parse"
)

func TestGetSize(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")
	fi := newFileInfos()

	err := fi.monitorFile()

	if err != nil {
		t.Fatal("Failed get file monitor data: error: %v", err)
	}
}
