package user

import (
	"fmt"
	"github.com/ExchangeProject/db"
	"github.com/ExchangeProject/mnemonic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New(c *gin.Context) {
	appid := c.Query("appid")
	appsecret := c.Query("appsecret")
	userid := c.Query("userid")
	coin := c.Query("coin")
	//判断appid和appsecret是否合法
	valid, err := db.AppIdIsValid(appid, appsecret)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "未知异常",
		})
		fmt.Println(err)
		return
	}
	if !valid {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "appip或appsecret非法",
		})
		fmt.Println(err)
		return
	}
	//判断当前用户是否存在
	exist, _mnemonic, err := db.IsExistByUserId(appid, userid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "未知异常",
		})
		fmt.Println(err)
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
			fmt.Println(err)
			return
		}
		err = db.InsertUser(appid, userid, _mnemonic)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "写入用户失败",
			})
			fmt.Println("InsertUserErr: ", err)
			return
		}
	} else {
		//判断代币地址是否存在
		exist, coinaddress, err := db.IsCoinAddressExist(appid, userid, coin)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "未知异常",
			})
			fmt.Println("InsertUserErr: ", err)
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
		fmt.Println("InsertUserErr: ", err)
		return
	}
	//m/44'/60'/0'/0
	path := fmt.Sprintf("m/44'/%d'/0'/0/0", coinindex)
	fmt.Println(coin, coinindex, path)
	var info mnemonic.AccountInfo
	if coin == "ETH" {
		info, err = mnemonic.GenerateAccount(_mnemonic, path)
	} else if coin == "BTC" {
		info, err = mnemonic.GenerateBtcAccount(_mnemonic, path)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "私匙生成失败",
		})
		fmt.Println("InsertUserErr: ", err)
		return
	}
	err = db.AddCoinAddress(appid, userid, coin, info.PrivateKey, info.Address)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "写入数据失败",
		})
		fmt.Println("InsertUserErr: ", err)
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
