package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// сделать   конфиг и валидацию
func main() {
	var user, hostname, password, dbname string
	var port int

	viper.AutomaticEnv() //это  нужно чтобы иметь  доступ к ENV  через viper
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println(viper.GetString("REST_USER"))
	if !viper.IsSet("REST_USER") {
		log.Fatal("missing REST_USER")
	} else {
		user = viper.GetString("REST_USER")
	}
	if !viper.IsSet("REST_PASSWORD") {
		log.Fatal("missing REST_PASSWORD")
	} else {
		password = viper.GetString("REST_PASSWORD")
	}
	if !viper.IsSet("REST_DBNAME") {
		log.Fatal("missing REST_DBNAME")
	} else {
		dbname = viper.GetString("REST_DBNAME")

	}
	if !viper.IsSet("REST_HOST") {
		log.Fatal("missing REST_HOST")
	} else {
		hostname = viper.GetString("REST_HOST")

	}
	if !viper.IsSet("REST_PORT") {
		log.Fatal("missing REST_PORT")
	} else {
		port = viper.GetInt("REST_PORT")

	}
	a := App{}
	//InitDb(user, password, dbname, host string, port int)* sql.DB, error
	//db, err := InitDb("rtk", "rtk", "rtk", "localhost", 5432)
	db, err := InitDb(user, password, dbname, hostname, port)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	a.Initialize(db)

	a.Run(":8080")

}
