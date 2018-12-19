package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/qizikd/ExchangeProject/settings"
)

func GetConn() (*sql.DB, error) {
	if settings.DBMysql != nil && settings.DBMysql.Ping() == nil {
		return settings.DBMysql, nil
	}
	db, err := sql.Open("mysql", settings.MySqlSourceName)
	if err != nil {
		glog.Error("连接mysql失败", settings.MySqlSourceName, err)
		return nil, err
	}
	settings.DBMysql = db
	return db, err
}
