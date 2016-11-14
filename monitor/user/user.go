package user

import (
	"fmt"
	"time"

	"monitor/alert"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/shirou/gopsutil/host"
)

type UserStat struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  int    `json:"started"`
}

type UserStats struct {
	AllUser []UserStat `json:"userinfos"`
}

// get all disk useage.

func newUserStats() *UserStats {
	return &UserStats{}
}

func GetUserStat() (*UserStats, error) {
	user := newUserStats()
	return user.getUserStat()
}

func (u *UserStats) getUserStat() (*UserStats, error) {

	users, err := host.Users()

	if err != nil {
		logx.FError("getUserStat to get user status error: %v", err)
		return &UserStats{}, err
	}

	for _, v := range users {
		u.AllUser = append(u.AllUser, UserStat{
			User:     v.User,
			Terminal: v.Terminal,
			Host:     v.Host,
			Started:  v.Started,
		})
	}
	return u, nil
}

func MonitorUsers() {
	go func() {
		user := newUserStats()
		logx.FInfo("%v", "Login module started!")
		if err := user.monitorUsers(); err != nil {
			logx.FCritical("MonitorUers error: %v", err)
		}
	}()
}

func (u *UserStats) monitorUsers() error {

	cf := parse.NewCf()

	if cf.UserConfig == nil {
		logx.FError("%v", "UserConfig is nil!")
		return fmt.Errorf("UserConfig is nil!")
	}

	//Clean slice
	u.reset()
	t1 := time.NewTicker(time.Second * time.Duration(cf.UserConfig.Lasttime))
	for {
		select {
		case <-t1.C:
			users, err := u.getUserStat()
			if err != nil {
				continue
			}

			var exist bool
			for _, v := range users.AllUser {
				logx.FDebug("Login info User: %v Terminal: %v Host: %v", v.User, v.Terminal, v.Host)
				exist = false
				for _, vv := range cf.UserConfig.Userlist {
					if v.User == vv {
						exist = true
						break
					}
				}
				if !exist {
					content := fmt.Sprintf("Alarm: Invalid user  %s! \nfrom: %v ", v.User, cf.Addr)
					timestamp := time.Now()
					header := " invalid username " + v.User
					alert.AlertConvergence(cf.UserConfig.EmailArray, timestamp, content, header)
				}
			}
			u.reset()
		}
	}
	return nil
}

func (u *UserStats) reset() {
	u.AllUser = u.AllUser[:0:0]
}
