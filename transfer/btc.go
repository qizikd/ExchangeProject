package transfer

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

type OmniApiTxRef struct {
	Address      string                                 `json:"address"`
	Pages        int                                    `json:"pages"`
	Transactions []rpcclient.Omni_ListtransactionResult `json:"transactions"`
}

var (
	gobcyToken     = "9184cf751ace44f090769b52643ade0b"
	gobcyChain     = "main"
	omnicoreHost   = "47.92.148.83:8332"
	omnicoreUser   = "omnicorerpc"
	omnicorePass   = "abcd1234"
	omnipropertyid = 31
)

func TransactionBtc(fromAddress string, toAddress string, privateKey string, amount int) (tx string, err error) {
	bcy := gobcy.API{gobcyToken, "btc", gobcyChain}
	//讲私匙从wif格式转换为原始格式
	privwif := privateKey
	privb, _, _ := base58.CheckDecode(privwif)
	privstr := hex.EncodeToString(privb)
	privstr = privstr[0 : len(privstr)-2]
	//Post New TXSkeleton
	trans := gobcy.TempNewTX(fromAddress, toAddress, amount)
	skel, err := bcy.NewTX(trans, false)
	//Sign it locally
	err = skel.Sign([]string{privstr})
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
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		//rpc.blockchain.info
		Host: omnicoreHost,
		User: omnicoreUser,
		Pass: omnicorePass,
	}, nil)
	if err != nil {
		glog.Error("error creating new btc client: ", err)
		return
	}
	defer client.Disconnect()
	tx, err = client.OmniSend(fromAddress, toAddress, strconv.FormatFloat(float64(amount)/btcutil.SatoshiPerBitcoin, 'f', 8, 64), omnipropertyid)
	if err != nil {
		glog.Error(err)
		return "", err
	}
	return
}

func GetBalanceBtc(address string) (balance int, err error) {
	bcy := gobcy.API{gobcyToken, "btc", gobcyChain}
	addr, err := bcy.GetAddrBal(address, nil)
	if err != nil {
		glog.Error(err)
		return
	}
	return addr.Balance, nil
}

func GetBalanceUSDT(address string) (balance int, err error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         omnicoreHost,
		User:         omnicoreUser,
		Pass:         omnicorePass,
	}, nil)
	if err != nil {
		glog.Error("error creating new btc client: ", err)
		return
	}
	defer client.Disconnect()
	balance, err = client.GetOmniBalance(address, omnipropertyid)
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

func GetBtcTransactions(address string) (addr gobcy.Addr, err error) {
	bcy := gobcy.API{gobcyToken, "btc", gobcyChain}
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
	var ops OmniApiTxRef
	err = json.Unmarshal(body, &ops)
	if err != nil {
		glog.Error("解析json字符串失败", err)
		return
	}
	return ops.Transactions, nil
}

/*
func GetUsdtTransactions(address string, count int, skip int) (result []rpcclient.Omni_ListtransactionResult, err error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         omnicoreHost,
		User:         omnicoreUser,
		Pass:         omnicorePass,
	}, nil)
	if err != nil {
		glog.Error("error creating new btc client: ", err)
		return
	}
	defer client.Disconnect()
	result, err = client.Omni_Listtransactions(address, count, skip)
	if err != nil {
		glog.Error(err)
		return
	}
	return result, err
}
*/
func ImportPrivkey(privkey string, label string) (err error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         omnicoreHost,
		User:         omnicoreUser,
		Pass:         omnicorePass,
	}, nil)
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
	return client.ImportPrivKeyRescan(wif, label, false)
}
