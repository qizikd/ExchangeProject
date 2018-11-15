package transfer

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ExchangeProject/token"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/glog"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
)

/*
type EthplorerAddTxs struct {
	Timestamp int     `json:"timestamp"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	TxHash    string  `json:"hash"`
	Value     float32 `json:"value"`
	Success   bool    `json:"success"`
}

type EthplorerHistoryTx struct {
	Timestamp int    `json:"timestamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	TxHash    string `json:"transactionHash"`
	Value     string `json:"value"`
	Type      string `json:"type"`
}

type EthplorerHistoryTxs struct {
	Operations []EthplorerHistoryTx `json:"operations"`
}

type EthplorerToken struct {
	Address     string `json:"address"`
	Name        string `json:"name"`
	Decimals    string `json:"decimals"`
	Symbol      string `json:"symbol"`
	TotalSupply string `json:"totalsupply"`
}

type EthplorerErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
*/

type EtherscanEthTx struct {
	Hash      string `json:"hash"`
	TimeStamp string `json:"timeStamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Gas       string `json:"gas"`
	GasPrice  string `json:"gasPrice"`
	GasUsed   string `json:"gasUsed"`
	IsError   string `json:"isError"`
}

type EtherscanEthResult struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Result  []EtherscanEthTx `json:"result"`
}

type EtherscanTokenTx struct {
	Hash      string `json:"hash"`
	TimeStamp string `json:"timeStamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Gas       string `json:"gas"`
	GasPrice  string `json:"gasPrice"`
	GasUsed   string `json:"gasUsed"`
}

type EtherscanTokenResult struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Result  []EtherscanTokenTx `json:"result"`
}

var ethplorerApiKey = "uvbw1547DAjgI32" //freekey
var etherscanApiKey = "Y4E31DPUZHG7XQDVQU31233GHMB98YBTIW"

func TransactionEth(toAddress string, privateKey string, amount int64) (tx string, err error) {
	tx, err = sendRawTransaction(privateKey, toAddress, nil, amount)
	if err != nil {
		glog.Error("sendRawTransaction", err)
		return
	}
	return
}

func TransactionToken(toAddress string, tokenAddress string, privateKey string, tokenAmount int64) (tx string, err error) {
	data := buildTransfer(toAddress, tokenAmount)
	tx, err = sendRawTransaction(privateKey, tokenAddress, data, 0)
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

func GetEthBalance(address string) (amount big.Int, err error) {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		glog.Error("连接infura节点失败", err)
		return
	}
	defer client.Close()
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		glog.Error("获取Eth余额失败", err)
		return
	}
	return *balance, nil
}

func GetTokenBalance(tokenAddress string, address string) (amount big.Int, err error) {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		glog.Error("连接infura节点失败", err)
		return
	}
	defer client.Close()
	instance, err := token.NewToken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		glog.Error(err)
		return
	}
	balance, err := instance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	if err != nil {
		glog.Error("获取token balance失败", err)
		return
	}
	return *balance, nil
}

func GetEthTransactions(address string, page int, offset int) (txs []EtherscanEthTx, err error) {
	//https://api.etherscan.io/api?module=account&action=txlist&address=0x16dC60B242E301c40541Fe89CA4065471dE12ba3&startblock=0&endblock=99999999&page=1&offset=10&sort=asc&apikey=YourApiKeyToken
	//url := fmt.Sprintf("http://api.ethplorer.io/getAddressTransactions/%s?apiKey=%s", address, ethplorerApiKey)
	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&startblock=0&endblock=99999999&address=%s&apikey=%s&page=%d&offset=%d&sort=desc",
		address, etherscanApiKey, page, offset)
	resp, err := http.Get(url)
	if err != nil {
		glog.Error("http请求失败", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = errors.New("获取交易失败")
		//fmt.Println("HTTP " + strconv.Itoa(resp.StatusCode) + " " + http.StatusText(resp.StatusCode))
		if resp.StatusCode == http.StatusForbidden {
			body, _ := ioutil.ReadAll(resp.Body)
			glog.Error(url, string(body))
			//fmt.Println(string(body))
		}
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return
	}
	var ops EtherscanEthResult
	err = json.Unmarshal(body, &ops)
	if err != nil {
		glog.Error("解析json字符串失败", err)
		return
	}
	if ops.Status == "0" {
		glog.Error("返回错误", url, ops.Message)
		err = errors.New(ops.Message)
		return
	}
	return ops.Result, nil
}

func GetTokenTransactions(tokenaddress string, address string, page int, offset int) (txs []EtherscanTokenTx, err error) {
	//https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=0x0f466615b79f8b8973734b4941ac26c8e995ee7c&address=0x16dC60B242E301c40541Fe89CA4065471dE12ba3&page=1&offset=10&sort=desc&apikey=YourApiKeyToken
	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&apikey=%s&page=%d&offset=%d&sort=desc",
		tokenaddress, address, etherscanApiKey, page, offset)
	//url := fmt.Sprintf("http://api.ethplorer.io/getAddressHistory/%s?apiKey=%s&token=%s&type=transfer", address, ethplorerApiKey, tokenaddress)
	resp, err := http.Get(url)
	if err != nil {
		glog.Error("http请求失败", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = errors.New("获取交易失败")
		fmt.Println("HTTP " + strconv.Itoa(resp.StatusCode) + " " + http.StatusText(resp.StatusCode))
		if resp.StatusCode == http.StatusForbidden {
			body, _ := ioutil.ReadAll(resp.Body)
			glog.Error(url, string(body))
		}
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return
	}
	var ops EtherscanTokenResult
	err = json.Unmarshal(body, &ops)
	if err != nil {
		glog.Error("解析json字符串失败", err)
		return
	}
	if ops.Status == "0" {
		glog.Error("返回错误", url, ops.Message)
		err = errors.New(ops.Message)
		return
	}
	return ops.Result, nil
}

func buildTransfer(toAddressHex string, tokenAmount int64) (data []byte) {
	toAddress := common.HexToAddress(toAddressHex)
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	//生成data methodABI+参数
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	amount := big.NewInt(tokenAmount)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return
}

func sendRawTransaction(privatekey string, toAddressHex string, data []byte, value int64) (tx string, err error) {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		glog.Error("连接infura节点失败", err)
		return
	}
	defer client.Close()
	//发送方私匙
	privateKey, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		glog.Error(err)
		return
	}
	//发送方公匙
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		glog.Error("error casting public key to ECDSA", err)
		return
	}
	//from地址
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//获取noce值
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		glog.Error(err)
		return
	}
	//获取gasPrice
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		glog.Error(err)
		return
	}
	//fmt.Println("gasPrice:", gasPrice)
	toAddress := common.HexToAddress(toAddressHex)

	//获取gasLimit
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     fromAddress,
		To:       &toAddress,
		GasPrice: gasPrice,
		Data:     data,
	})
	if err != nil {
		glog.Error(err)
	}

	//gasLimit := uint64(300000)
	//fmt.Printf("gasLimit:%d\n", gasLimit) // 23256
	//创建一个交易对象
	ethvalue := big.NewInt(value) // in wei
	transaction := types.NewTransaction(nonce, toAddress, ethvalue, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		glog.Error(err)
		return
	}

	signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		glog.Error(err)
		return
	}

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	//rawTxHex := hex.EncodeToString(rawTxBytes)
	//fmt.Printf(rawTxHex) // f86...772

	tx2 := new(types.Transaction)
	rlp.DecodeBytes(rawTxBytes, &tx2)

	err = client.SendTransaction(context.Background(), tx2)
	if err != nil {
		glog.Error(err)
		return
	}
	return tx2.Hash().Hex(), nil
}
