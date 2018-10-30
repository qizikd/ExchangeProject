package db

import (
	"ExchangeProject/core/mysql"
)

func InsertUser(appid string,userid string,mnemonic string) (err error) {
	conn , err := mysql.GetConn()
	if err != nil {
		return
	}
	stmt, err := conn.Prepare(`INSERT user (appid, userid, mnemonic) VALUES (?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(appid,userid,mnemonic)
	return
}

func IsExistByUserId(appid string,userid string) (exist bool,mnemonic string,err error) {
	conn , err := mysql.GetConn()
	if err != nil {
		 return
	}
	stmt , err := conn.Prepare("select mnemonic from user where appid=? and userid=?")
	if err != nil {
		return
	}
	rows, err := stmt.Query(appid,userid)
	if err != nil {
		return
	}
	exist = rows.Next()
	if exist{
		rows.Scan(&mnemonic)
	}
	return
}