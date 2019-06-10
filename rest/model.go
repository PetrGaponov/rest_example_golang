package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	namesUrl = "http://country.io/names.json"
	phoneUrl = "http://country.io/phone.json"
)

type SQLDatabase interface { //this  functions may contains DB specific SQL query
	UpdateData(db *sql.DB) error
	FindCountry(db *sql.DB, countryName string) (string, error)
}

type PostgresDB struct {
}

//we can  mock  this  func in tests for db access
func InitDbPostgres(user, password, dbname, host string, port int) (*sql.DB, error) {
	//Function  that  return  db  connect  to specific database
	dbinfo := fmt.Sprintf("user=%s host=%s  port=%d  password=%s dbname=%s sslmode=disable",
		user, host, port, password, dbname)
	db, err := sql.Open("postgres", dbinfo) // our pull Postgres concurently safe!
	if err != nil {
		return nil, errors.Wrap(errors.New("Can not open database"), err.Error())
	}
	db.SetConnMaxLifetime(10 * time.Minute) //сколько живут сессии
	db.SetMaxIdleConns(11)                  // максимальное колличество  готовых  соединений
	err = db.Ping()                         //check connect
	if err != nil {
		return nil, errors.Wrap(errors.New("Can not connect to database"), err.Error())
	}

	return db, nil

}

//
func (*PostgresDB) UpdateData(db *sql.DB) error {
	// update    names

	countryMap, err := GetRequest(namesUrl)
	if err != nil {
		log.Println(errors.Cause(err))
		return err
	}

	names := make([]string, 0, len(countryMap))
	for key := range countryMap {
		names = append(names, key)
	}
	sort.Strings(names) //sort by key

	for _, key := range names {

		if _, ok := countryMap[key].(string); ok {
			//log.Println("index : ", strings.ToLower(key), " value : ", strings.ToLower(countryMap[key].(string)))
			_, err = db.Exec("INSERT INTO names (country_letter, country_name) VALUES ($1, $2) ON CONFLICT (country_letter) DO UPDATE SET  country_letter = EXCLUDED.country_letter, country_name = EXCLUDED.country_name", key, strings.ToLower(countryMap[key].(string)))
			if err != nil {
				log.Println(errors.Cause(err))
				return err
			}
		}
	}
	// update    phone
	names = names[:0]

	phoneMap, err := GetRequest(phoneUrl)
	if err != nil {
		log.Println(errors.Cause(err))
		return err
	}

	for key := range phoneMap {
		names = append(names, key)
	}
	sort.Strings(names) //sort by key

	for _, key := range names {
		if _, ok := phoneMap[key].(string); ok {
			//log.Println("index : ", strings.ToLower(key), " value : ", strings.ToLower(phoneMap[key].(string)))
			_, err = db.Exec("INSERT INTO phone (country_letter, country_code) VALUES ($1, $2) ON CONFLICT (country_letter) DO UPDATE SET  country_letter = EXCLUDED.country_letter, country_code = EXCLUDED.country_code", key, strings.ToLower(phoneMap[key].(string)))
			if err != nil {
				log.Println(errors.Cause(err))
				return err
			}
		}
	}

	return err
	//end update
}

func (*PostgresDB) FindCountry(db *sql.DB, countryName string) (string, error) {
	//var answer CountryCode
	var countryCode string
	sqlStatement := `select country_code from phone join names on names.country_letter=phone.country_letter where names.country_name=$1`
	row := db.QueryRow(sqlStatement, countryName)
	err := row.Scan(&countryCode)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return "", errors.Wrap(errors.New("Result is empty"), err.Error())
	case nil:
		return countryCode, nil
	default:
		log.Println("Can not make SQL response")
		//log.Println(err.Error())
		return "", errors.Wrap(errors.New("Can not make SQL request"), err.Error())
	}

}
