package transfer

import (
	"encoding/hex"
	"fmt"
	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
)

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
