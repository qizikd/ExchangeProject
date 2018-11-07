package transfer

import (
	"encoding/hex"
	"fmt"
	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
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

func TransactionBtc(fromAddress string, toAddress string, privateKey string, amount int) (tx string, err error) {
	bcy := gobcy.API{"9184cf751ace44f090769b52643ade0b", "btc", "test3"}
	//讲私匙从wif格式转换为原始格式
	privwif := privateKey
	privb, _, _ := base58.CheckDecode(privwif)
	privstr := hex.EncodeToString(privb)
	privstr = privstr[0 : len(privstr)-2]
	//Post New TXSkeleton
	skel, err := bcy.NewTX(gobcy.TempNewTX(fromAddress, toAddress, amount), false)
	//Sign it locally
	err = skel.Sign([]string{privstr})
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//Send TXSkeleton
	skel, err = bcy.SendTX(skel)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return skel.Trans.Hash, nil
}

func GetBalanceBtc(address string) (balance int, err error) {
	bcy := gobcy.API{"9184cf751ace44f090769b52643ade0b", "btc", "test3"}
	addr, err := bcy.GetAddrBal(address, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	return addr.Balance, nil
}

func GetBtcTransactions(address string) (addr gobcy.Addr, err error) {
	bcy := gobcy.API{"9184cf751ace44f090769b52643ade0b", "btc", "test3"}
	addr, err = bcy.GetAddrFull(address, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
