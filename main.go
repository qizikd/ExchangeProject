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
	router.GET("/user/sendto", user.SendTo)
	router.GET("/user/balance", user.Balance)
	router.GET("/user/balanceEthToken", user.BalanceToken)
	router.GET("/user/tokenTxs", user.TokenTransactions)
	router.GET("/user/ethTxs", user.EthTransactions)
	router.Run(":" + *port)
}
