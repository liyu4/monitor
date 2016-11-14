package alert

import (
	"testing"
)

func TestSendMail(t *testing.T) {
	sendEmail([]string{"kevin.chen@digitalx.cn"}, "test", "test")
}
