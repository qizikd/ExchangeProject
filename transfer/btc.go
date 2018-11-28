package transfer

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ExchangeProject/settings"
	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/golang/glog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type GobcyAddInfo struct {
	Address             string       `json:"address"`
	Total_received      int          `json:"total_received"`
	Total_sent          int          `json:"total_sent"`
	Balance             int          `json:"balance"`
	Unconfirmed_balance int          `json:"unconfirmed_balance"`
	Final_balance       int          `json:"final_balance"`
	N_tx                int          `json:"n_tx"`
	Unconfirmed_n_tx    int          `json:"unconfirmed_n_tx"`
	Final_n_tx          int          `json:"final_n_tx"`
	TxRefs              []GobcyTxRef `json:"txrefs"`
}

type GobcyTxRef struct {
	Block_height  int    `json:"block_height"`
	Tx_hash       string `json:"tx_hash"`
	Tx_input_n    int    `json:"tx_input_n"`
	Tx_output_n   int    `json:"tx_output_n"`
	Value         int    `json:"value"`
	Spent         bool   `json:"spent"`
	Double_spend  bool   `json:"double_spend"`
	Confirmations int    `json:"confirmations"`
	Ref_balance   int    `json:"ref_balance"`
	Confirmed     string `json:"confirmed"`
	Received      string `json:"received"`
}

func newGobcy() (api gobcy.API) {
	//随机获取一个GobcyApi key，防止用一个有请求限制
	gobcyApis := []string{"9184cf751ace44f090769b52643ade0b", "269d9eb40f3048a6875b45e5aee017e9", "64cb8fe59b934d8d9df104fa8d59a85b",
		"3dc9ad5c6d8449a499103de610ab12d8", "c2c26b546bf04f049ea06e7e539d868a", "dd4b3e08a28347dfa222469472ad1a73"}
	content, err := ioutil.ReadFile("gobcyApiToken.json")
	if err == nil {
		json.Unmarshal(content, &gobcyApis)
	}
	rand.Seed(time.Now().Unix())
	apiIndex := rand.Intn(len(gobcyApis))
	var gobcyApi = gobcyApis[apiIndex]
	if settings.IsBTCTestNet3 {
		api = gobcy.API{gobcyApi, "btc", "test3"}
	} else {
		api = gobcy.API{gobcyApi, "btc", "main"}
	}
	return
}

func newOmniClient() (client *rpcclient.Client, err error) {
	if settings.IsBTCTestNet3 {
		client, err = rpcclient.New(&rpcclient.ConnConfig{
			HTTPPostMode: true,
			DisableTLS:   true,
			//rpc.blockchain.info
			Host: "39.104.156.29:18332",
			User: "omnicorerpc",
			Pass: "abcd1234",
		}, nil)
	} else {
		client, err = rpcclient.New(&rpcclient.ConnConfig{
			HTTPPostMode: true,
			DisableTLS:   true,
			//rpc.blockchain.info
			Host: "47.92.148.83:8332",
			User: "omnicorerpc",
			Pass: "abcd1234",
		}, nil)
	}
	return
}

func TransactionBtc(fromAddress string, toAddress string, privateKey string, amount int) (tx string, err error) {
	bcy := newGobcy()
	//讲私匙从wif格式转换为原始格式
	privwif := privateKey
	privb, _, err := base58.CheckDecode(privwif)
	if err != nil {
		glog.Error(err)
		return "", err
	}
	privstr := hex.EncodeToString(privb)
	privstr = privstr[0 : len(privstr)-2]
	//Post New TXSkeleton
	trans := gobcy.TempNewTX(fromAddress, toAddress, amount)
	skel, err := bcy.NewTX(trans, false)
	//Sign it locally
	var priv []string
	for i := 0; i < len(skel.ToSign); i++ {
		priv = append(priv, privstr)
	}
	err = skel.Sign(priv)
	if err != nil {
		glog.Error(err)
		return "", err
	}
	//Send TXSkeleton
	skel, err = bcy.SendTX(skel)
	if err != nil {
		glog.Error(err)
		return "", err
	}
	return skel.Trans.Hash, nil
}

func TransactionUsdt(fromAddress string, toAddress string, privateKey string, amount int) (tx string, err error) {
	url := fmt.Sprintf("http://39.104.156.29:8333/user/sendto?privkey=%s&fromaddress=%s&toaddress=%s&amount=%d",
		privateKey, fromAddress, toAddress, amount)
	resp, err := http.Get(url)
	if err != nil {
		glog.Error("http请求失败", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = errors.New("获取失败")
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Error(url, string(body))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return
	}
	type sendResult struct {
		Errcode int    `json:"errcode"`
		Txid    string `json:"txid"`
		Msg     string `json:"msg"`
	}
	var result sendResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		glog.Error("解析json字符串失败", err)
		return
	}
	if result.Errcode != 0 {
		err = errors.New("发送交易失败")
		glog.Error("发送交易失败", result.Msg)
		return
	}
	return result.Txid, nil
}

func GetBalanceBtc(address string) (balance int, err error) {
	bcy := newGobcy()
	addr, err := bcy.GetAddrBal(address, nil)
	if err != nil {
		glog.Error(err)
		return
	}
	return addr.Balance, nil
}

func GetBalanceUSDT(address string) (balance int, err error) {
	client, err := newOmniClient()
	if err != nil {
		glog.Error("error creating new btc client: ", err)
		return
	}
	defer client.Disconnect()
	if settings.IsBTCTestNet3 {
		balance, err = client.GetOmniBalance(address, 1)
	} else {
		balance, err = client.GetOmniBalance(address, 31)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

/*
func GetBalanceUSDT(address string) (balance int, err error) {
	url := fmt.Sprintf("http://39.104.156.29:8333/user/balance?address=%s", address)
	resp, err := http.Get(url)
	if err != nil {
		glog.Error("http请求失败", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = errors.New("获取失败")
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Error(url, string(body))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return
	}
	type balanceResult struct {
		Errcode int `json:"errcode"`
		Amount  int `json:"amount"`
	}
	var result balanceResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		glog.Error("解析json字符串失败", err)
		return
	}
	if result.Errcode != 0 {
		err = errors.New("余额查询失败")
		return
	}
	return result.Amount, nil
}
*/

func GetBtcTransactions(address string) (addr gobcy.Addr, err error) {
	bcy := newGobcy()
	addr, err = bcy.GetAddrFull(address, nil)
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

func GetUsdtTransactions(address string, page int) (result []rpcclient.Omni_ListtransactionResult, err error) {
	url := "https://api.omniexplorer.info/v1/transaction/address"
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(fmt.Sprintf("addr=%s&page=%d", address, page)))
	if err != nil {
		glog.Error("http请求失败", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = errors.New("获取交易失败")
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Error(url, string(body))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return
	}
	type OmniApiTxRef struct {
		Address      string                                 `json:"address"`
		Pages        int                                    `json:"pages"`
		Transactions []rpcclient.Omni_ListtransactionResult `json:"transactions"`
	}
	var ops OmniApiTxRef
	err = json.Unmarshal(body, &ops)
	if err != nil {
		glog.Error("解析json字符串失败", err)
		return
	}
	return ops.Transactions, nil
}

func ImportPrivkey(privkey string, label string) (err error) {
	client, err := newOmniClient()
	if err != nil {
		glog.Error("error creating new btc client: ", err)
		return
	}
	defer client.Disconnect()
	wif, err := btcutil.DecodeWIF(privkey)
	if err != nil {
		glog.Error(err)
		return
	}
	fmt.Println(time.Now(), "私钥导入开始")
	err = client.ImportPrivKeyRescan(wif, label, false)
	fmt.Println(time.Now(), "私钥导入结束")
	return err
}
