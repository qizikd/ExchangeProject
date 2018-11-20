package user

import (
	"fmt"
	"github.com/ExchangeProject/db"
	"github.com/ExchangeProject/mnemonic"
	"github.com/ExchangeProject/settings"
	"github.com/ExchangeProject/transfer"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"strconv"
)

func VerifyAppId(appid string, appsecret string) bool {
	//判断appid和appsecret是否合法
	valid, err := db.AppIdIsValid(appid, appsecret)
	if err != nil {
		glog.Error(err)
		return false
	}
	if !valid {
		glog.Error("appip或appsecret非法")
		return false
	}
	return true
}

func New(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	userid := c.Query("userid")
	coin := c.Query("coin")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	//判断当前用户是否存在
	exist, _mnemonic, err := db.IsExistByUserId(appid, userid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "未知异常",
		})
		glog.Error(err)
		return
	}
	//不存在就生成助记词
	if !exist {
		_mnemonic, err = mnemonic.GenerateMnemonic()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "生成助记词失败",
			})
			glog.Error(err)
			return
		}
		err = db.InsertUser(appid, userid, _mnemonic)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "写入用户失败",
			})
			glog.Error(err)
			return
		}
	} else {
		//判断代币地址是否存在
		exist, coinaddress, _, err := db.IsCoinAddressExist(appid, userid, coin)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "未知异常",
			})
			glog.Error(err)
			return
		}
		//如果存在直接返回
		if exist {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"data": gin.H{
					"coin":    coin,
					"address": coinaddress,
				},
			})
			return
		}
	}
	//生成代币地址
	coinindex, err := db.GetCoinIndex(coin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "当前数字货币不支持",
		})
		glog.Error(err)
		return
	}
	//m/44'/60'/0'/0
	path := fmt.Sprintf("m/44'/%d'/0'/0/0", coinindex)
	var info mnemonic.AccountInfo
	if coin == "ETH" {
		info, err = mnemonic.GenerateAccount(_mnemonic, path)
	} else if coin == "BTC" {
		if settings.IsBTCTestNet3 {
			path = "m/44'/1'/1'/0/0"
		}
		info, err = mnemonic.GenerateBtcAccount(_mnemonic, path)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "地址生成失败",
		})
		glog.Error(err)
		return
	}
	err = db.AddCoinAddress(appid, userid, coin, info.PrivateKey, info.Address)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "写入数据失败",
		})
		glog.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"coin":    coin,
			"address": info.Address,
		},
	})
}

func SendTo(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	userid := c.Query("fromuserid")
	coin := c.Query("coin")
	toAddress := c.Query("toaddress")
	_amount := c.Query("amount")
	amount, err := strconv.Atoi(_amount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "金额非法",
		})
		return
	}
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	//判断当前用户是否存在
	exist, _, err := db.IsExistByUserId(appid, userid)
	if err != nil || !exist {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户不存在",
		})
		if err != nil {
			glog.Error(err)
		}
		return
	}
	//判断代币地址和私匙是否存在
	coinSymbol := coin
	if coin == "ETH" || coin == "ERC-20" {
		coinSymbol = "ETH"
	} else {
		coinSymbol = "BTC"
	}
	exist, fromAddress, privKey, err := db.IsCoinAddressExist(appid, userid, coinSymbol)
	if err != nil || !exist {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "代币地址不存在",
		})
		if err != nil {
			glog.Error(err)
		}
		return
	}
	switch coin {
	case "BTC":
		txHex, err := transfer.TransactionBtc(fromAddress, toAddress, privKey, amount)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "交易失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"txHex": txHex,
			"txUrl": fmt.Sprintf("https://live.blockcypher.com/btc-testnet/tx/%s/", txHex),
		})
		return
	case "USDT":
		txHex, err := transfer.TransactionUsdt(fromAddress, toAddress, privKey, amount)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "交易失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"txHex": txHex,
			"txUrl": fmt.Sprintf("https://live.blockcypher.com/btc-testnet/tx/%s/", txHex),
		})
		return
	case "ETH":
		txHex, err := transfer.TransactionEth(toAddress, privKey, int64(amount))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "交易失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"txHex": txHex,
			"txUrl": fmt.Sprintf("https://etherscan.io/tx/%s", txHex),
		})
		return
	case "ERC-20":
		tokenAddress := c.Query("tokenaddress")
		txHex, err := transfer.TransactionToken(toAddress, tokenAddress, privKey, int64(amount))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "交易失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"txHex": txHex,
			"txUrl": fmt.Sprintf("https://etherscan.io/tx/%s", txHex),
		})
		return
	}
}

func Balance(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	coin := c.Query("coin")
	address := c.Query("address")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	switch coin {
	case "BTC":
		balance, err := transfer.GetBalanceBtc(address)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "获取余额失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"coin":    coin,
				"address": address,
				"balance": balance,
			},
		})
	case "ETH":
		balance, err := transfer.GetEthBalance(address)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "获取余额失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"coin":    coin,
				"address": address,
				"balance": balance.String(),
			},
		})
		return
	case "USDT":
		balance, err := transfer.GetBalanceUSDT(address)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "获取余额失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"coin":    coin,
				"address": address,
				"balance": strconv.Itoa(balance),
			},
		})
		return
	default:
		return
	}
}

func BalanceToken(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	tokenAddress := c.Query("tokenaddress")
	address := c.Query("address")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	balance, err := transfer.GetTokenBalance(tokenAddress, address)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取余额失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"address": address,
			"balance": balance.String(),
		},
	})
}

func TokenTransactions(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	tokenAddress := c.Query("tokenaddress")
	address := c.Query("address")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	p := c.Query("page")
	page, err := strconv.Atoi(p)
	if err != nil {
		page = 1
	}
	s := c.Query("pagesize")
	offset, err := strconv.Atoi(s)
	if err != nil {
		offset = 10
	}
	txs, err := transfer.GetTokenTransactions(tokenAddress, address, page, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取交易历史失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"address":      address,
			"tokenaddress": tokenAddress,
			"txs":          txs,
		},
	})
}

func EthTransactions(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	address := c.Query("address")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	p := c.Query("page")
	page, err := strconv.Atoi(p)
	if err != nil {
		page = 1
	}
	s := c.Query("pagesize")
	offset, err := strconv.Atoi(s)
	if err != nil {
		offset = 10
	}
	txs, err := transfer.GetEthTransactions(address, page, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取交易历史失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"address": address,
			"txs":     txs,
		},
	})
}

func BtcTransactions(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	address := c.Query("address")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}

	txs, err := transfer.GetBtcTransactions(address)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取交易历史失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"address": address,
			"txs":     txs.TXs,
		},
	})
}

func UsdtTransactions(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	address := c.Query("address")
	//判断appid和appsecret是否合法
	if !VerifyAppId(appid, appsecret) {
		c.JSON(http.StatusOK, gin.H{
			"code": -333,
			"msg":  "appip或appsecret非法",
		})
		return
	}
	p := c.Query("page")
	page, err := strconv.Atoi(p)
	if err != nil || page == 0 {
		page = 1
	}
	//omnicore page是从0开始
	page = page - 1
	txs, err := transfer.GetUsdtTransactions(address, page)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取交易历史失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"address": address,
			"txs":     txs,
		},
	})
}
