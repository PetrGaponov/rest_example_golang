package main

// сделать   конфиг и валидацию
func main() {
	a := App{}
	//InitDb(user, password, dbname, host string, port int)* sql.DB, error
	db, err := InitDb("rtk", "rtk", "rtk", "localhost", 5432)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	a.Initialize(db)

	a.Run(":8080")

}
