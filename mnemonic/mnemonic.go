package mnemonic

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/qizikd/ExchangeProject/settings"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"net/http"
)

type AccountInfo struct {
	PrivateKey string
	PulickKey  string
	Address    string
}

func NewMnemonic(c *gin.Context) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		glog.Error("生成助记词失败", err)
		return
	}
	c.JSON(200, gin.H{
		"error":    0,
		"mnemonic": mnemonic,
	})
}

func NewAccounts(c *gin.Context) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		result := gin.H{
			"errcode": -1,
			"msg":     err.Error(),
		}
		glog.Error(err)
		c.JSON(http.StatusOK, result)
		return
	}
	generateAccounts(c, mnemonic, "m/44'/60'/0'/0/0")
}

func MoreAccounts(c *gin.Context) {
	mnemonic := c.Query("mnemonic")
	if len(mnemonic) == 0 {
		result := gin.H{
			"errcode": -1,
			"msg":     "mnemonic 不允许为空",
		}
		c.JSON(http.StatusOK, result)
		return
	}
	path := c.Query("path")
	if len(path) == 0 {
		path = "m/44'/60'/0'/0/0"
	}
	generateAccounts(c, mnemonic, path)
}

func generateAccounts(c *gin.Context, mnemonic string, p string) {
	defer func() {
		if err := recover(); err != nil {
			result := gin.H{
				"errcode": 1,
				"msg":     "path 无效",
			}
			glog.Error(result)
			c.JSON(http.StatusOK, result)
		}
	}()

	seed := bip39.NewSeed(mnemonic, "")
	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		result := gin.H{
			"errcode": -2,
			"msg":     err.Error(),
		}
		glog.Error(err)
		c.JSON(http.StatusOK, result)
		return
	}

	path := hdwallet.MustParseDerivationPath(p)
	account, err := wallet.Derive(path, false)
	if err != nil {
		result := gin.H{
			"errcode": -3,
			"msg":     err.Error(),
		}
		glog.Error(err)
		c.JSON(http.StatusOK, result)
		return
	}
	publickeyHex, err := wallet.PublicKeyHex(account)
	if err != nil {
		result := gin.H{
			"errcode": -4,
			"msg":     err.Error(),
		}
		glog.Error(err)
		c.JSON(http.StatusOK, result)
		return
	}
	privatekeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		result := gin.H{
			"errcode": -5,
			"msg":     err.Error(),
		}
		glog.Error(err)
		c.JSON(http.StatusOK, result)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errcode": 0,
		"data": gin.H{
			"mnemonic":   mnemonic,
			"privatekey": privatekeyHex,
			"publickey":  publickeyHex,
			"address":    account.Address.Hex(),
		},
	})
}

func GenerateAccount(mnemonic string, p string) (myAccount AccountInfo, err error) {
	fmt.Printf("mnemonic: %s,path: %s\n", mnemonic, p)

	defer func() {
		if err2 := recover(); err2 != nil {
			glog.Error(err2)
			err = errors.New("未知错误")
		}
	}()

	seed := bip39.NewSeed(mnemonic, "")

	btckey, err := bip32.NewMasterKey(seed)

	fmt.Println(btckey.B58Serialize(), btckey.ChainCode, hex.EncodeToString(btckey.Key))

	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		glog.Error(err)
		return
	}

	path := hdwallet.MustParseDerivationPath(p)
	account, err := wallet.Derive(path, true)
	if err != nil {
		glog.Error(err)
		return
	}

	myAccount.PulickKey, err = wallet.PublicKeyHex(account)
	if err != nil {
		glog.Error(err)
		return
	}
	myAccount.PrivateKey, err = wallet.PrivateKeyHex(account)
	if err != nil {
		glog.Error(err)
		return
	}
	myAccount.Address = account.Address.Hex()
	return
}

func GenerateBtcAccount(mnemonic string, p string) (myAccount AccountInfo, err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			glog.Error(err2)
			err = errors.New("未知错误")
		}
	}()

	seed := bip39.NewSeed(mnemonic, "")
	var extkey *hdkeychain.ExtendedKey
	if settings.IsBTCTestNet3 {
		extkey, err = hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
	} else {
		extkey, err = hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	}

	//根据extkey计算对应path下的key
	path := hdwallet.MustParseDerivationPath(p)
	child := extkey
	for _, n := range path {
		child, err = child.Child(n)
		if err != nil {
			glog.Error(err)
			return
		}
	}
	//account_pub, err := child.Neuter()
	//if err != nil {
	//	return
	//}
	//fmt.Println("Derived PrivateKey: ", child.String())
	//fmt.Println("Derived PublicKey: ", account_pub.String())
	//wif
	private_key, err := child.ECPrivKey()
	if err != nil {
		glog.Error(err)
		return
	}
	var private_wif *btcutil.WIF
	if settings.IsBTCTestNet3 {
		private_wif, err = btcutil.NewWIF(private_key, &chaincfg.TestNet3Params, true)
	} else {
		private_wif, err = btcutil.NewWIF(private_key, &chaincfg.MainNetParams, true)
	}

	if err != nil {
		glog.Error(err)
		return
	}
	var address_key *btcutil.AddressPubKeyHash
	if settings.IsBTCTestNet3 {
		address_key, err = child.Address(&chaincfg.TestNet3Params)
	} else {
		address_key, err = child.Address(&chaincfg.MainNetParams)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	//private_str := private_wif.String()
	//address_str := address_key.String()
	//fmt.Println("private_str: ", private_str, "address_str: ", address_str)

	myAccount.PulickKey = ""
	myAccount.PrivateKey = private_wif.String()
	myAccount.Address = address_key.String()
	//fmt.Printf("Address: %s,Public Key: %s, Private Key: %s\n", myAccount.Address, myAccount.PulickKey, myAccount.PrivateKey)
	return
}

func GenerateMnemonic() (mnemonic string, err error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		glog.Error(err)
		return
	}
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		glog.Error(err)
		return
	}
	return
}
