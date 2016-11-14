package alert

import (
	"monitor/queue"
	"strings"
	"sync"
	"time"

	gomail "github.com/gopkg.in/gomail.v2"
	"github.com/kevinchen/logx"
)

type convergenceData struct {
	scrollStartTime         int64
	scrollLeft, scrollRight int64
}

var (
	convergence      = make(map[string]*convergenceData)
	convergenceMutex = sync.RWMutex{}
	stack            = queue.InitQueue()
)

const (
	convergenceMinInterval int64 = 60 * 5
	convergenceMapLimit    int   = 300
	convergenceMaxInterval int64 = 60 * 40
)

func AlertConvergence(mailArray []string, timestamp time.Time, content, header string) {
	go alertConvergence(mailArray, timestamp, content, header)
}

func alertConvergence(mailArray []string, timestamp time.Time, content, header string) {
	convergenceMutex.Lock()
	cod := convergence[header]

	// init convergenceData
	{
		if nil == cod {
			if nil == convergence[header] {
				cod = &convergenceData{scrollStartTime: timestamp.Unix(), scrollLeft: 0, scrollRight: 1}
				convergence[header] = cod
				if len(convergence) > convergenceMapLimit {
					clearConvergenceMap()
				}
			}
		}
	}

	//convergence and update scroll if need
	duration := (timestamp.Unix() - cod.scrollStartTime)
	repeatAlter := duration < cod.scrollLeft*convergenceMinInterval
	if repeatAlter {
		convergenceMutex.Unlock()
		return
	}
	cod.scroll(timestamp)
	convergenceMutex.Unlock()

	// Second convergence.

	convergenceMutex.Lock()
	data := queue.Node{
		Mail:    mailArray,
		Content: content,
		Header:  header,
	}
	push(data)
	convergenceMutex.Unlock()
	//send alert
}

func push(data queue.Node) {
	queue.EnQueue(stack, data)
}

func pop() queue.Node {
	node := queue.DeQueue(stack)
	return node
}

func (cod *convergenceData) scroll(timestamp time.Time) {
	reset := (timestamp.Unix() - cod.scrollStartTime) > cod.scrollRight*convergenceMinInterval
	if reset {
		cod.scrollLeft, cod.scrollRight, cod.scrollStartTime = 0, 1, timestamp.Unix()
		return
	}

	reachMax := cod.scrollRight*convergenceMinInterval >= convergenceMaxInterval
	if reachMax {
		cod.scrollLeft, cod.scrollRight = cod.scrollRight, cod.scrollRight+(cod.scrollRight-cod.scrollLeft)
	} else {
		cod.scrollLeft, cod.scrollRight = cod.scrollRight, cod.scrollRight+(cod.scrollRight-cod.scrollLeft)*2
	}

}

func clearConvergenceMap() {
	haveExpired := false
	var minStartTime int64
	minStartTimeKey := ""
	expires := make([]string, 0)
	for k, v := range convergence {
		// get expired array and oldest time

		if time.Now().Unix()-v.scrollStartTime > v.scrollLeft*convergenceMinInterval {
			haveExpired = true
			expires = append(expires, k)
		}

		if 0 == minStartTime || v.scrollStartTime < minStartTime {
			minStartTime = v.scrollStartTime
			minStartTimeKey = k
		}

	}
	// delete expired or oldest time if need
	if haveExpired {
		for _, k := range expires {
			delete(convergence, k)
		}
	} else {
		delete(convergence, minStartTimeKey)
	}
}

var timer = time.NewTicker(time.Minute * 1)

func MialQueueTask() {
	logx.FInfo("%v", "Mail queue task stared!")
	go func() {
		for {
			select {
			case <-timer.C:
				aggregation := merger()
				if aggregation == nil {
					continue
				}

				for k, v := range aggregation {
					mailArray := strings.Split(k, " ")
					sendEmail(mailArray, v, "Aggregation alert!")
				}

			}
		}
	}()
}

func alldata() []queue.Node {
	list := make([]queue.Node, 0)
	convergenceMutex.Lock()
	for {
		data := pop()
		if data.Mail == nil {
			break
		}
		list = append(list, data)
		if data.Mail == nil {
			break
		}
	}
	convergenceMutex.Unlock()
	return list
}

func merger() map[string]string {
	all := alldata()

	if all == nil {
		return nil
	}

	arrange := make(map[string]string, 100)

	for i := 0; i < len(all); i++ {
		mailstring := strings.Join(all[i].Mail, " ")

		if value, ok := arrange[mailstring]; ok {
			arrange[mailstring] = value + all[i].Content + "\n"
		} else {
			arrange[mailstring] = all[i].Content + "\n"
		}
	}
	return arrange
}

func sendEmail(mailArray []string, content, header string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "test@digitalx.cn")
	m.SetHeader("To", mailArray...)
	m.SetHeader("Subject", header)
	m.SetBody("text/plain", content)

	// Setting mail agent.
	// Notic:
	// Please input token instead of password.
	d := gomail.NewDialer("smtp.exmail.qq.com", 465, "test@digitalx.cn", "123456Test")

	maxRetry := 3
	retry := maxRetry
	var err error

	for retry > 0 {
		retry--
		err = d.DialAndSend(m)

		if err == nil {
			break
		} else {
			time.Sleep(time.Second * 5)
		}
	}

	if retry == 0 {
		logx.FError("dail tcp error: %v", err)
		return
	}

	logx.FInfo("%v", "send mail is success")
}

func sendEmailUseHtml(mailArray []string, content, header string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "test@digitalx.cn")
	m.SetHeader("To", mailArray...)
	m.SetHeader("Subject", header)
	m.SetBody("text/html", content)

	// Setting mail agent.
	// Notic:
	// Please input token instead of password.
	d := gomail.NewDialer("smtp.exmail.qq.com", 465, "test@digitalx.cn", "123456Test")

	maxRetry := 3
	retry := maxRetry
	var err error

	for retry > 0 {
		retry--
		err = d.DialAndSend(m)

		if err == nil {
			break
		} else {
			time.Sleep(time.Second * 5)
		}
	}

	if retry == 0 {
		logx.FError("dail tcp error: %v", err)
		return
	}

	logx.FInfo("%v", "send mail is success")
}

// OnlySend is Send email.
func OnlySend(mailArray []string, content, header string) {
	sendEmailUseHtml(mailArray, content, header)
}
