package main

import (
	"./ttt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()
	
	//RESTFul 路由
	router.GET("/hello-world",helloWorld)
	router.POST("/",helloWorldPost)

	//启动服务
	_ = router.Run("127.0.0.1:8080")
}

//RESTFul路由 Get函数
func helloWorld(c *gin.Context)  {
	ttt.Abs(http.StatusOK,"Enter",c)
}

//RESTFul路由 post函数
func helloWorldPost(c *gin.Context)  {
	c.String(http.StatusOK,"Hello World in POST")
}
