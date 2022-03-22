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
	MSG = "Sorry i did not understand that"
)

func Alpha(w http.ResponseWriter, r *http.Request) {
	jsonData := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err == nil {
		if question, ok := jsonData["text"].(string); ok {
			if answer, code, err := Service(question); err == nil {
				jsonResponse := map[string]interface{}{"text": answer}
				w.WriteHeader(code)
				json.NewEncoder(w).Encode(jsonResponse)
			} else {
				http.Error(w, err.Error(), code)
			}
		} else {
			http.Error(w, "Error getting text field from JSON", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
	}
}

func Service(question string) (string, int, error) {
	client := &http.Client{}
	uri := URI + "?appid=" + KEY + "&i=" + url.QueryEscape(question)
	if req, err := http.NewRequest("GET", uri, nil); err == nil {
		if rsp, err := client.Do(req); err == nil {
			if rsp.StatusCode == http.StatusOK {
				if body, err := ioutil.ReadAll(rsp.Body); err == nil {
					answer := string(body)
					return answer, http.StatusOK, nil
				}
				//Return error message for each status code
			} else if rsp.StatusCode == http.StatusNotImplemented {
				//A 501 error is returned if the question is not understood, hence return a misunderstanding message rather than an error
				return MSG, http.StatusOK, nil
			} else if rsp.StatusCode == http.StatusForbidden {
				return "", http.StatusForbidden, errors.New("error in response from Wolfram Alpha. appid missing or invalid")
			} else {
				return "", rsp.StatusCode, errors.New("error in response from Wolfram Alpha service")
			}
		}
	}
	return "", http.StatusInternalServerError, errors.New("error making requst to Wolfram Alpha")
}

func main() {
	//Create Router and listen for POST requests on localhost:3001/alpha
	r := mux.NewRouter()
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	http.ListenAndServe(":3001", r)
}
