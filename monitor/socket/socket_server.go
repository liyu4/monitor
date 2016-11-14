package socket

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"monitor/alert"
	"monitor/parse"
	"monitor/protocol"

	"github.com/kevinchen/logx"
	"github.com/mygojson"
)

// Json data returned to the client.
type SendToClient struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var record = make(map[string]interface{}, 0)

func Run() {
	func() {
		go server()
	}()
}

func server() {
	var cf = parse.NewCf()

	if cf.SocketConfig == nil {
		logx.FError("%v", "parse socketconfig failed in socket server!")
		return
	}
	netListen, err := net.Listen("tcp", cf.SocketConfig.Addr+":"+cf.SocketConfig.Port)
	checkError(err)
	defer netListen.Close()

	logx.FInfo("%v", "Socket server staring...! ")

	for {
		conn, err := netListen.Accept()
		if err != nil {
			logx.FError("server listening failed: %v", err)
		}
		logx.FInfo("%v", conn.RemoteAddr().String()+" tcp connect success! ")
		host := conn.RemoteAddr().String()
		go handleConnection(conn, host)
	}
}

func response(conn net.Conn) {
	sendtoclient := SendToClient{Code: 200, Msg: "THE SERVER IS RUNING! "}
	response, err := json.Marshal(sendtoclient)
	if err != nil {
		logx.FError("response failed:  %v", err.Error())
	}
	packet := protocol.Packet(response)
	conn.Write(packet)
}

func handleConnection(conn net.Conn, host string) {
	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)

	//声明一个管道用于接收解包的数据
	readerChannel := make(chan []byte, 16)
	go reader(readerChannel, host)
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			logx.FError("%v connection error: %v", conn.RemoteAddr().String(), err.Error())
			return
		}

		// Parsing the received packet.
		tmpBuffer = protocol.Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
		response(conn)
	}
}

func reader(readerChannel chan []byte, host string) {
	var cf = parse.NewCf()

	for {
		select {
		case data := <-readerChannel:
			js := mygojson.Json(string(data)).Get("all")
			array, err := js.Array()

			//Parse data error.
			if err != nil {
				logx.FError("parse json data failed!")
				continue
			}

			if len(array) == 0 {
				content := "Remote process information is blank , check that the socket-client program  whether is running , or the network whether is normal!"
				header := "Remote process is empty!"
				timestamp := time.Now()
				alert.AlertConvergence(cf.SocketConfig.EmailArray, timestamp, content, header)
			}

			var hasWrong bool = false

			for _, v := range array {
				if value, ok := v.(map[string]interface{}); ok {
					if rawname, ok := value["modulename"]; ok {
						for k, vv := range cf.CommonConfig.ModuleName {
							if rawname == vv {
								break
							} else if k+1 == len(cf.CommonConfig.ModuleName) {
								hasWrong = true
								content := "Process " + rawname.(string) + "  is wrong" + "\n" + "it at least one of local monitor list that is: " + strings.Join(cf.CommonConfig.ModuleName, "-") + "\n" + "from socket: " + host
								header := fmt.Sprintf("Wrong process name is %v and remoteHost is %v", rawname, host)
								timestamp := time.Now()
								alert.AlertConvergence(cf.SocketConfig.EmailArray, timestamp, content, header)
							}
						}
					}
				}
			}

			if hasWrong {
				continue
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
								content := fmt.Sprintf("%v have lost or closed, receive program process id is %v\nError from socket: %v ", modulename, pid, host)
								header := fmt.Sprintf("%v have lost,from specified host %v", modulename, host)
								timestamp := time.Now()
								alert.AlertConvergence(mail, timestamp, content, header)
							}
						}
					}
				}
			}

		//Timeout
		case <-time.After(time.Second * time.Duration(cf.SocketConfig.Timeout)):
			content := "Client maybe lost, please check out socket-client" + "\n" +
				"whether online or ping" + "\n" + host

			header := "Client no response" + host
			timestamp := time.Now()
			alert.AlertConvergence(cf.SocketConfig.EmailArray, timestamp, content, header)
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		logx.FCritical("%v Fatal error: %v", os.Stderr, err.Error())
	}
}
