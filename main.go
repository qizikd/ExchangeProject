package main

import (
	"ExchangeProject/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main()  {
	router := gin.Default()
	router.GET("/", func(c *gin.Context){
		c.String(http.StatusOK, "Hello, World!")
	})
	router.GET("/user/create", user.New)
	router.Run(":81")
}
