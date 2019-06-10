package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	_ "github.com/lib/pq"
)

//эти переменные  можно  тоже убрать внутрь приложения
//render.Render(w, r, NewAnswerWithId("Call is successfuly queued", GetID(r.Context())))
var tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
var myClient = &http.Client{Timeout: 2 * time.Second, Transport: tr}
var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

//

type Answer struct { // answer with json tel code
	TelCode        string `json:"tel_code"`
	HTTPStatusCode int    `json:"-"` // http response status code
}

//
type SuccessResponse struct { //success response for  update data
	StatusText     string `json:"status"`
	HTTPStatusCode int    `json:"-"` // http response status code
}

//
func (a *Answer) Render(w http.ResponseWriter, r *http.Request) error { //
	render.Status(r, a.HTTPStatusCode)
	return nil
}

func (a *SuccessResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, a.HTTPStatusCode)
	return nil
}

//

type App struct { //our  APP. Also  we can save our global logger here
	Model  SQLDatabase
	Router *chi.Mux
	DB     *sql.DB
}
type ErrResponse struct {
	Err            error  `json:"-"`               // low-level runtime error
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	AppCode        int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func (a *App) Initialize(db *sql.DB) {

	//var err error
	a.DB = db
	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger) //log our  request
	a.Router.Use(middleware.Recoverer)
	a.Router.Use(render.SetContentType(render.ContentTypeJSON))

	a.Router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, r, ErrNotFound) //if exist not  valid symbols, return not found
	})
	//
	a.Router.Route("/code/{country}", func(r chi.Router) { //
		r.Mount("/", GetRouter(a))
	})

	a.Router.Route("/reload", func(r chi.Router) {
		r.Mount("/", PostRouter(a))
	})

}

func GetRouter(a *App) http.Handler {
	router := chi.NewRouter()
	router.Get("/", a.GetCountry)
	return router
}

func PostRouter(a *App) http.Handler {
	router := chi.NewRouter()
	router.Post("/", a.ReloadData)
	return router
}

func (a *App) GetCountry(w http.ResponseWriter, r *http.Request) { //

	//
	var answer Answer
	//
	space := regexp.MustCompile(`\s+`) //remove  double  spaces  (may be wrong?)
	//

	if country := strings.ToLower(chi.URLParam(r, "country")); country != "" {
		// Make  validate param
		re := regexp.MustCompile(`^[a-zA-Z. ]*$`) //valid symbols
		if !re.MatchString(country) {
			render.Render(w, r, ErrNotFound) //if exist not  valid symbols, return not found
			return
		}
		country = space.ReplaceAllString(country, " ")
		res, err := a.Model.FindCountry(a.DB, country)
		if err != nil || res == "" { // for some countries in database (South Georgia and the South Sandwich Islands) code is empty ("") we shell return not found/
			render.Render(w, r, ErrNotFound) //
			return
		} else {
			answer.TelCode = res
			answer.HTTPStatusCode = 200
			render.Render(w, r, &answer) //
		}

	} else {
		render.Render(w, r, ErrNotFound) //
		return
	}

}

func (a *App) ReloadData(w http.ResponseWriter, r *http.Request) { //
	var answer = SuccessResponse{"Success", 200}
	//response, _ := json.Marshal(answer)
	err := a.Model.UpdateData(a.DB)
	if err == nil {
		render.Render(w, r, &answer) //
		//return
	} else {
		answer = SuccessResponse{"Error while update", 500}
		fmt.Println(errors.Cause(err))
		fmt.Println(err)
		render.Render(w, r, &answer) //
	}

}

func GetRequest(url string) (map[string]interface{}, error) { // список стран или телефонов
	var body []byte

	var netClient = &http.Client{
		Timeout: time.Second * 2,
	}

	response, err := netClient.Get(url)

	if err != nil { //response  may be nill. First check error
		return nil, err
	}
	defer response.Body.Close()
	//
	if response.StatusCode != 200 {

		return nil, errors.Wrap(errors.New("Error response code: "), strconv.Itoa(response.StatusCode))
	} else {

		body, err = ioutil.ReadAll(response.Body)

		if err != nil {
			return nil, errors.Wrap(errors.New("Error while read body"), err.Error())
		}
	}
	result := make(map[string]interface{})

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, errors.Wrap(errors.New("Error while unmarshal result"), err.Error())
	}
	return result, nil
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
