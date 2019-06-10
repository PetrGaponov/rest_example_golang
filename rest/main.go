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

	v, err := checkConfig()
	if err != nil {
		panic(err)
	}
	user = v.GetString("REST_USER")
	password = v.GetString("REST_PASSWORD")
	dbname = v.GetString("REST_DBNAME")
	hostname = v.GetString("REST_HOST")
	port = v.GetInt("REST_PORT")

	//InitDb(user, password, dbname, host string, port int)* sql.DB, error
	//db, err := InitDb("rtk", "rtk", "rtk", "localhost", 5432)
	pg := &PostgresDB{}
	db, err := InitDbPostgres(user, password, dbname, hostname, port)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	a := App{Model: pg}
	//a.Model = pg
	a.Initialize(db)
	log.Println("Loading data...")
	err = pg.UpdateData(a.DB) //load data
	if err != nil {
		log.Fatal("Can not load data!", err)
	} else {
		log.Println("Success load data")
	}

	log.Println("Ready for requests")
	a.Run(":8080")

}

//check config func
func checkConfig() (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv() //for access to ENV
	//
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println(v.GetString("REST_USER"))
	if !v.IsSet("REST_USER") {
		log.Fatal("missing REST_USER")
	}
	if !v.IsSet("REST_PASSWORD") {
		log.Fatal("missing REST_PASSWORD")
	}
	if !v.IsSet("REST_DBNAME") {
		log.Fatal("missing REST_DBNAME")
	}
	if !v.IsSet("REST_HOST") {
		log.Fatal("missing REST_HOST")
	}
	if !v.IsSet("REST_PORT") {
		log.Fatal("missing REST_PORT")
	}

	return v, err
}
