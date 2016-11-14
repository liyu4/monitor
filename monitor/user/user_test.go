package user

import (
	"testing"

	"monitor/parse"
)

func TestGet(t *testing.T) {
	parse.UpdateCf("/Users/admin/svn/src/monitor/conf/app.conf")
	u := newUserStats()

	err := u.monitorUsers()

	if err != nil {
		t.Fatal("Failed get users monitor data! error: %v", err)
	}
}
