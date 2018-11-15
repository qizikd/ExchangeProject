package transfer

import (
	"github.com/ExchangeProject/db"
	"github.com/ethereum/go-ethereum/log"
	"time"
)

func SyncImportPrivkey() {
	for {
		users, err := db.GetNoImportUserBtcPrivkey()
		if err != nil {
			log.Error("GetNoImportUserBtcPrivkey：", err)
			time.Sleep(10 * time.Second)
			continue
		}
		for key, value := range users {
			err := ImportPrivkey(value, key)
			if err != nil {
				log.Error("ImportPrivkey: ", err)
				continue
			}
			//
			err = db.UpdateImported(key)
			if err != nil {
				log.Error("UpdateImported: ", err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
