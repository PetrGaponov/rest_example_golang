package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
)

type AnyString struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyString) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

type FakeDB struct {
}

//
func (*FakeDB) UpdateData(db *sql.DB) error {
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
			mock.ExpectExec("INSERT INTO names").WithArgs(AnyString{}, AnyString{}).WillReturnResult(sqlmock.NewResult(1, 1))
			_, err = db.Exec("INSERT INTO names (country_letter, country_name) VALUES ($1, $2) ON CONFLICT (country_letter) DO UPDATE SET  country_letter = EXCLUDED.country_letter,	 country_name = EXCLUDED.country_name", key, strings.ToLower(countryMap[key].(string)))
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
			mock.ExpectExec("INSERT INTO phone").WithArgs(AnyString{}, AnyString{}).WillReturnResult(sqlmock.NewResult(1, 1))
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

func (*FakeDB) FindCountry(db *sql.DB, countryName string) (string, error) {
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

var a App
var mock sqlmock.Sqlmock
var rows_ru *sqlmock.Rows
var rows_us *sqlmock.Rows
var db *sql.DB
var err error

func TestMain(m *testing.M) {
	fakeDB := &FakeDB{}

	//db, err := InitDb("rtk", "rtk", "rtk", "localhost", 5432)
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("can not opendatabase")
	}
	a = App{Model: fakeDB}
	defer db.Close()
	if err != nil {
		panic(err)
	}
	//expected  rows for  GET Country test
	rows_ru = sqlmock.NewRows([]string{"country_code"}).AddRow("7")
	rows_us = sqlmock.NewRows([]string{"country_code"}).AddRow("1")

	a.Initialize(db)

	code := m.Run()

	os.Exit(code)
}

func TestGetCountry(t *testing.T) {
	mock.ExpectQuery("^select country_code from phone join names on names.country_letter=phone.country_letter").WithArgs("russia").WillReturnRows(rows_ru)

	req, _ := http.NewRequest("GET", "/code/russia", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["tel_code"] != "7" {
		t.Errorf("Expected tel_code  to be '7'. Got '%v'", m["tel_code"])
	}

}

//
func TestNotFound(t *testing.T) {
	mock.ExpectQuery("^select country_code from phone join names on names.country_letter=phone.country_letter").WithArgs("mordor").WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest("GET", "/code/Mordor", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["status"] != "Resource not found." {
		t.Errorf("Expected status  to be 'Resource not found.'. Got '%v'", m["status"])
	}

}

func TestCapitalLetter(t *testing.T) {
	mock.ExpectQuery("^select country_code from phone join names on names.country_letter=phone.country_letter").WithArgs("united states").WillReturnRows(rows_us)

	req, _ := http.NewRequest("GET", "/code/United States", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["tel_code"] != "1" {
		t.Errorf("Expected tel_code  to be '1'. Got '%v'", m["tel_code"])
	}

}

//

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestUpdate(t *testing.T) {

	req, _ := http.NewRequest("POST", "/reload", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["status"] != "Success" {
		t.Errorf("Expected status  to be 'Success'. Got '%v'", m["status"])
	}

}
