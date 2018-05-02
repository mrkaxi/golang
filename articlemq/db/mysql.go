package db

import (
	"database/sql"
	"log"
	//调用mysql的init方法
	_ "github.com/go-sql-driver/mysql"
	"github.com/widuu/goini"
)

//SQLDB 链接变量
var SQLDB *sql.DB

func init() {
	var err error
	conf := goini.SetConfig("config.ini")
	host := conf.GetValue("db", "host")
	port := conf.GetValue("db", "port")
	user := conf.GetValue("db", "user")
	password := conf.GetValue("db", "password")
	database := conf.GetValue("db", "database")

	SQLDB, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/"+database+"?parseTime=true")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = SQLDB.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}
}
