package mysql

import
(
	"ExchangeProject/settings"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

func GetConn()  (*sql.DB,error){
	if settings.DBMysql != nil && settings.DBMysql.Ping() == nil {
		return 	settings.DBMysql,nil
	}
	db, err := sql.Open("mysql", settings.MySqlSourceName)
	return db,err
}