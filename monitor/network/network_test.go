package network

import (
	"testing"

	"monitor/parse"
)

func TestGetRxAndTx(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")

	net := newNetwork()

	err := net.monitorNetwork()

	if err != nil {
		t.Fatal("Failed get network monitor data! error: %v", err)
	}
}
