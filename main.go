package main

import (
	"flag"
	"github.com/ExchangeProject/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	port := flag.String("port", "80", "Listen port")
	flag.Parse()
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})
	router.GET("/user/create", user.New)
	router.Run(":" + *port)
}
