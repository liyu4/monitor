package routers

import (
	"github.com/astaxie/beego"
	"monitor/controllers"
)

func init() {
	beego.Router("/", &controllers.Monitor{}, "post:Server")
}
