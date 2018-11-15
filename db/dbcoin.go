package db

import (
	"errors"
	"fmt"
	"github.com/ExchangeProject/core/mysql"
)

func AddCoinAddress(appid string, userid string, coin string, privatekey string, address string) (err error) {
	exist, _, _, err := IsCoinAddressExist(appid, userid, coin)
	if err != nil || exist {
		return
	}
	conn, err := mysql.GetConn()
	if err != nil {
		return
	}
	stmt, err := conn.Prepare(`INSERT coinaddress (appid, userid, coin, privatekey, address) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(appid, userid, coin, privatekey, address)
	if err != nil {
		return
	}
	return
}

func IsCoinAddressExist(appid string, userid string, coin string) (exist bool, coinaddress string, privatekey string, err error) {
	conn, err := mysql.GetConn()
	if err != nil {
		return
	}
	stmt, err := conn.Prepare("select address,privatekey from coinaddress where appid=? and userid=? and coin=?")
	if err != nil {
		return
	}
	rows, err := stmt.Query(appid, userid, coin)
	if err != nil {
		return
	}
	exist = rows.Next()
	if exist {
		rows.Scan(&coinaddress, &privatekey)
	}
	return
}

func GetCoinIndex(coin string) (coinindex int, err error) {
	conn, err := mysql.GetConn()
	if err != nil {
		return
	}
	stmt, err := conn.Prepare("select coinindex from coin where coin=?")
	if err != nil {
		return
	}
	rows, err := stmt.Query(coin)
	if err != nil {
		return
	}

	if rows.Next() {
		rows.Scan(&coinindex)
	} else {
		err = errors.New("当前数字货币不支持")
	}
	return
}

func GetNoImportUserBtcPrivkey() (users map[string]string, err error) {
	conn, err := mysql.GetConn()
	if err != nil {
		return
	}
	stmt, err := conn.Prepare("select userid,privatekey from coinaddress where coin='BTC' and import=0")
	if err != nil {
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		return
	}
	var userid, privkey string
	users = make(map[string]string)
	for rows.Next() {
		rows.Scan(&userid, &privkey)
		users[userid] = privkey
	}
	return
}

func UpdateImported(userid string) (err error) {
	conn, err := mysql.GetConn()
	if err != nil {
		return
	}
	_, err = conn.Exec(fmt.Sprintf("update coinaddress set import = 1 where coin='BTC' and userid='%s'", userid))
	return
}
