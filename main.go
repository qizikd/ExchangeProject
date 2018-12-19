package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/qizikd/ExchangeProject/db"
	"github.com/qizikd/ExchangeProject/user"
	"github.com/qizikd/wallet/transfer"
	"net/http"
	"time"
)

func main() {
	port := flag.String("port", "80", "Listen port")
	flag.Parse()
	//go transfer.SyncImportPrivkey()
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
	router.GET("/user/btcTxs", user.BtcTransactions)
	router.GET("/user/usdtTxs", user.UsdtTransactions)
	router.Run(":" + *port)
}

func SyncImportPrivkey() {
	for {
		users, err := db.GetNoImportUserBtcPrivkey()
		if err != nil {
			glog.Error(err)
			time.Sleep(10 * time.Second)
			continue
		}
		for key, value := range users {
			err := transfer.ImportPrivkey(value, key)
			if err != nil {
				glog.Error(err)
				continue
			}
			//
			err = db.UpdateImported(key)
			if err != nil {
				glog.Error(err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
