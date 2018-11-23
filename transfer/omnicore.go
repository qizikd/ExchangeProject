package transfer

import (
	"github.com/ExchangeProject/db"
	"github.com/golang/glog"
	"time"
)

func SyncImportPrivkey() {
	for {
		users, err := db.GetNoImportUserBtcPrivkey()
		if err != nil {
			glog.Error(err)
			time.Sleep(10 * time.Second)
			continue
		}
		for key, value := range users {
			err := ImportPrivkey(value, key)
			if err != nil {
				glog.Error(err)
				continue
			}
			//
			err = db.UpdateImported(key)
			if err != nil {
				glog.Error(err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
