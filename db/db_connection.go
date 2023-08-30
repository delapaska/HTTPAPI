package db

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/delapaska/AvitoTest/internal/config"
	_ "github.com/lib/pq"
)

var (
	pathConf string
)

func init() {
	flag.StringVar(&pathConf, pathConf, "config/conf.yaml", "path to configuration file")
	flag.Parse()
}

func DBConnect() *sql.DB {
	config, err := config.NewConfig(pathConf)
	if err != nil {
		panic(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	
	return db
}
