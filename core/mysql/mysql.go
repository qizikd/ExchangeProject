package mysql

import (
	"database/sql"
	"github.com/ExchangeProject/settings"
	"github.com/ethereum/go-ethereum/log"
	_ "github.com/go-sql-driver/mysql"
)

func GetConn() (*sql.DB, error) {
	if settings.DBMysql != nil && settings.DBMysql.Ping() == nil {
		return settings.DBMysql, nil
	}
	db, err := sql.Open("mysql", settings.MySqlSourceName)
	if err != nil {
		log.Error("连接mysql失败", settings.MySqlSourceName, err)
		return nil, err
	}
	settings.DBMysql = db
	return db, err
}
