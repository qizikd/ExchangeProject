package db

import (
	"github.com/qizikd/ExchangeProject/core/mysql"
)

func AppIdIsValid(id string, secret string) (valid bool, err error) {
	conn, err := mysql.GetConn()
	if err != nil {
		return
	}
	stmt, err := conn.Prepare("select appid from appstore where appid=? and appsecret=?")
	if err != nil {
		return
	}
	rows, err := stmt.Query(id, secret)
	if err != nil {
		return
	}
	return rows.Next(), nil
}
