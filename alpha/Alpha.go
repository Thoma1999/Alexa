package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

const (
	URI = "http://api.wolframalpha.com/v1/result"
	KEY = "H5U44G-KRU4Q2TLR7"
)

func Alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if question, ok := t["text"].(string); ok {
			if answer, err := Service(question); err == nil {
				u := map[string]interface{}{"text": answer}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(u)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func Service(question string) (string, error) {
	client := &http.Client{}
	uri := URI + "?appid=" + KEY + "&i=" + url.QueryEscape(question)
	if req, err := http.NewRequest("GET", uri, nil); err == nil {
		if rsp, err := client.Do(req); err == nil {
			if rsp.StatusCode == http.StatusOK {
				if body, err := ioutil.ReadAll(rsp.Body); err == nil {
					answer := string(body)
					return answer, nil
				}
			}
		}
	}
	return "", errors.New("Service")
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	http.ListenAndServe(":3001", r)
}
