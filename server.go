package main

import (
	"fakedns/server"
	"fakedns/web"
	"github.com/gin-gonic/gin"
)

func main() {

	server.StartServer()
	//web 控制台
	r := gin.Default()
	web.Initroute(r)
	r.Run(":2333")
}
