package web

import "github.com/gin-gonic/gin"

func Initroute(r *gin.Engine) {
	//r.POST("/login", )
	//r.Static("/static", "./static")
	r.GET("/ping", Ping)
	v1 := r.Group("/dns") //TODO basic鉴权
	{
		v1.GET("/Command", Command)
		v1.GET("/GetTasks", GetTasks)
		v1.GET("/GetResults", GetResults)
		v1.GET("/GetClients", GetClients)

	}

}
