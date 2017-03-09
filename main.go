package main

import (
	"./tweet_queue/publisher/lib/dynamo_db_manager"
	"encoding/json"
	"fmt"
	"github.com/guregu/dynamo"
	"net/http"
	"reflect"
	"text/template"
)

type Result struct {
	Values        []dynamo_db_manager.LLMockTweetTrans
	LatestRageKey string
	RequestLang   string
}

const RETRIVE_LIMIT = 20
const HASH_KEY_NAME = "Lang"
const RANGE_KEY_NAME = "Timestamp_TweetID"

func handler_index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Fprintf(w, "Template Excute Error: ")
		fmt.Fprintf(w, err.Error())
		return
	}
}

func handler_api_init(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	ddm := dynamo_db_manager.New()
	var results []dynamo_db_manager.LLMockTweetTrans
	err := ddm.Table.
		Get(HASH_KEY_NAME, lang).
		//Range(RANGE_KEY_NAME, dynamo.Greater, "1488993166_12345678900").
		Order(dynamo.Descending).
		Limit(RETRIVE_LIMIT).
		All(&results)
	if err != nil {
		fmt.Fprintf(w, "DB Error: ")
		fmt.Fprintf(w, err.Error())
		return
	}

	rl := reflect.Indirect(reflect.ValueOf(results[0]))
	lrk := rl.FieldByName(RANGE_KEY_NAME).String()
	re := Result{
		results,
		lrk,
		lang,
	}
	js, err := json.Marshal(re)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func handler_api_get(w http.ResponseWriter, r *http.Request) {
	lrk := r.URL.Query().Get("lrk")
	lang := r.URL.Query().Get("lang")
	ddm := dynamo_db_manager.New()
	var results []dynamo_db_manager.LLMockTweetTrans
	err := ddm.Table.
		Get("Lang", lang).
		Range(RANGE_KEY_NAME, dynamo.Greater, lrk).
		//Order(dynamo.Descending).
		Limit(RETRIVE_LIMIT).
		All(&results)
	if err != nil {
		fmt.Fprintf(w, "DB Error: ")
		fmt.Fprintf(w, err.Error())
		return
	}
	js, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", handler_index)
	http.HandleFunc("/api/init", handler_api_init)
	http.HandleFunc("/api/get", handler_api_get)
	http.ListenAndServe(":8080", nil)
}
