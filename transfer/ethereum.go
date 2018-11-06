package transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ExchangeProject/token"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
)

func TransactionEth(toAddress string, privateKey string, amount int64) (tx string, err error) {
	tx, err = sendRawTransaction(privateKey, toAddress, nil, amount)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func TransactionToken(toAddress string, tokenAddress string, privateKey string, tokenAmount int64) (tx string, err error) {
	data := buildTransfer(toAddress, tokenAmount)
	tx, err = sendRawTransaction(privateKey, tokenAddress, data, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetEthBalance(address string) (amount big.Int, err error) {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		log.Error("连接infura节点失败", err)
		fmt.Println(err)
		return
	}
	defer client.Close()
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		log.Error("获取余额失败", err)
		fmt.Println(err)
		return
	}
	return *balance, nil
}

func GetTokenBalance(tokenAddress string, address string) (amount big.Int, err error) {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		log.Error("连接infura节点失败", err)
		return
	}
	defer client.Close()
	instance, err := token.NewToken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		log.Error("初始化token实例失败", err)
		return
	}
	balance, err := instance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	if err != nil {
		log.Error("获取token balance失败", err)
		return
	}
	fmt.Println("balance: ", balance)
	return *balance, nil
}

func buildTransfer(toAddressHex string, tokenAmount int64) (data []byte) {
	toAddress := common.HexToAddress(toAddressHex)
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println("methodID:", hexutil.Encode(methodID)) // 0xa9059cbb
	//生成data methodABI+参数
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println("toAddress:", hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d
	amount := big.NewInt(tokenAmount)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println("Amount", hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return
}

func sendRawTransaction(privatekey string, toAddressHex string, data []byte, value int64) (tx string, err error) {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		log.Error("连接infura节点失败", err)
		return
	}
	defer client.Close()
	//发送方私匙
	privateKey, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Error("私钥转换失败", err)
		return
	}
	//发送方公匙
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("error casting public key to ECDSA", err)
		return
	}
	//from地址
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//获取noce值
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Error("获取noce值失败", err)
		return
	}
	//获取gasPrice
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("获取gasprice失败", err)
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
		log.Error("获取gaslimit失败", err)
	}

	//gasLimit := uint64(300000)
	//fmt.Printf("gasLimit:%d\n", gasLimit) // 23256
	//创建一个交易对象
	ethvalue := big.NewInt(value) // in wei
	transaction := types.NewTransaction(nonce, toAddress, ethvalue, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Error("获取chainID失败", err)
		return
	}

	signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Error("签名交易失败", err)
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
		log.Error("发送交易失败", err)
		return
	}
	return tx2.Hash().Hex(), nil
}
