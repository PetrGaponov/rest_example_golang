package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	//db, err := InitDb("rtk", "rtk", "rtk", "localhost", 5432)
	db, mock, err := sqlmock.New()
	if err != nil {
		panic("can not opendatabase")
	}
	defer db.Close()
	if err != nil {
		panic(err)
	}

	rows := sqlmock.NewRows([]string{"country_code"}).AddRow("7")

	//mock.ExpectQuery("^select country_code from phone join names on names.country_letter=phone.country_letter where names.country_name='russia'$").WillReturnResult('7')
	mock.ExpectQuery("^select country_code from phone join names on names.country_letter=phone.country_letter").WithArgs("russia").WillReturnRows(rows)
	mock.ExpectQuery("^select country_code from phone join names on names.country_letter=phone.country_letter").WithArgs("mordor").WillReturnError(sql.ErrNoRows)
	a.Initialize(db)

	code := m.Run()

	os.Exit(code)
}

func TestGetCountry(t *testing.T) {

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

	req, _ := http.NewRequest("GET", "/code/Mordor", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["status"] != "Resource not found." {
		t.Errorf("Expected status  to be 'Resource not found.'. Got '%v'", m["status"])
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
