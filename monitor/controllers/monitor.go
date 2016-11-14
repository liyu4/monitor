package controllers

import (
	"fmt"
	"io/ioutil"
	"monitor/alert"
	"monitor/parse"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/kevinchen/logx"
	"github.com/mygojson"
)

type Monitor struct {
	beego.Controller
}

type HttpSendToClient struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var (
	lock   = sync.RWMutex{}
	record = make(map[string]int64, 300)
)

// Server Return to client
func (m *Monitor) Server() {
	var cf = parse.NewCf()

	if cf.HttpConfig != nil {
		if cf.HttpConfig.Sync {
			resp := HttpSendToClient{
				Code: 505,
				Msg:  "Server has shoudown!",
			}
			m.Data["json"] = resp
			m.ServeJSON()
			return
		}
	}

	body := m.Ctx.Request.Body

	request, err := ioutil.ReadAll(body)
	body.Close()

	if err != nil {
		logx.FError("Server error: %v", err)
		return
	}

	js := mygojson.Json(string(request)).Get("all")
	array, err := js.Array()

	if err != nil {
		logx.FError("parse json data error!")
	}

	var hasWrong bool = false
	var host = m.Ctx.Request.RemoteAddr

	if len(array) == 0 {
		content := "Remote process information is blank , check that the http-client program whether is running , or the network whether is normal!"
		header := "Remote process is empty!"
		timestamp := time.Now()
		alert.AlertConvergence(cf.SocketConfig.EmailArray, timestamp, content, header)
	}

	for _, v := range array {
		if value, ok := v.(map[string]interface{}); ok {
			if rawname, ok := value["modulename"]; ok {
				for k, vv := range cf.CommonConfig.ModuleName {
					if rawname == vv {
						break
					} else if k+1 == len(cf.CommonConfig.ModuleName) {
						hasWrong = true
						content := "Remote process " + rawname.(string) + " was wrong" + "\n" + "local monitor list is: " + strings.Join(cf.CommonConfig.ModuleName, "-") + "\n" + "from http: " + host
						header := fmt.Sprintf("Wrong process name is %v and remoteHost is %v (HTTP)", rawname, host)
						timestamp := time.Now()
						alert.AlertConvergence(cf.HttpConfig.EmailArray, timestamp, content, header)
					}
				}
			}
		}
	}

	if hasWrong {
		goto Leave
	}

	for _, v := range array {
		if value, ok := v.(map[string]interface{}); ok {
			intermoduleName := value["modulename"]
			interPid := value["pid"]
			modulename := intermoduleName.(string)
			pid := interPid.(float64)
			if pid == 0 {
				for _, vv := range cf.ModuleConfig.Ms {
					if modulename == vv.Name {
						mail := vv.EmailArray
						content := fmt.Sprintf("%v have lost or closed, receive program process id is %v\nError from http: %v ", modulename, pid, host)
						header := fmt.Sprintf("%v have lost,from specified host %v", modulename, host)
						timestamp := time.Now()
						alert.AlertConvergence(mail, timestamp, content, header)
					}
				}
			}
		}
	}

Leave:
	lock.Lock()
	record[m.Ctx.Request.RemoteAddr] = time.Now().Unix()
	lock.Unlock()
	resp := HttpSendToClient{
		Code: 200,
		Msg:  "The Server is running!",
	}
	m.Data["json"] = resp
	m.ServeJSON()
	return
}

func Isoff() {
	go func() {
		isOff()
	}()
}

func isOff() {
	var cf = parse.NewCf()

	timeout := cf.HttpConfig.Timeout
	Timer := time.NewTicker(time.Second * time.Duration(timeout))
	for {
		select {
		case <-Timer.C:
			lock.Lock()
			for addr, timestamp := range record {
				if time.Now().Unix()-timestamp >= cf.HttpConfig.Timeout {
					content := "Client maybe lost, please check out http-client whether online or ping " + addr
					header := "Client no response" + addr
					timestamp := time.Now()
					alert.AlertConvergence(cf.HttpConfig.EmailArray, timestamp, content, header)
					clear(addr)
				}
			}
			lock.Unlock()
		}
	}
}

func clear(flag string) {
	delete(record, flag)
}
