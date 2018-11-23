package settings

import "database/sql"

var (
	MySqlSourceName = "exchange:sSfwRTwp4Lnp5CM2@tcp(39.104.156.29:3306)/exchange"
	DBMysql         *sql.DB
	IsBTCTestNet3   = true
)
