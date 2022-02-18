package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Stt(question interface{}) (map[string]interface{}, error) {
	client := &http.Client{}
	uri := "http://localhost:3002/stt"
	if jsonspeech, err := json.Marshal(question); err == nil {
		if req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonspeech)); err == nil {
			if rsp, err := client.Do(req); err == nil {
				if rsp.StatusCode == http.StatusOK {
					t := map[string]interface{}{}
					if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
						if q, ok := t["text"].(string); ok {
							fmt.Println(q)
						}
						return t, nil
					}
				} else {
					fmt.Println(rsp.StatusCode)
				}
			}
		}
	}
	return nil, errors.New("Error converting speech to text")
}

func Alpha(question interface{}) (map[string]interface{}, error) {
	client := &http.Client{}
	uri := "http://localhost:3001/alpha"
	if jsontext, err := json.Marshal(question); err == nil {
		fmt.Println(string(jsontext))
		if req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsontext)); err == nil {
			if rsp, err := client.Do(req); err == nil {
				if rsp.StatusCode == http.StatusOK {
					t := map[string]interface{}{}
					if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
						if a, ok := t["text"].(string); ok {
							fmt.Println(a)
						}
						return t, nil
					}
				} else {
					fmt.Println(rsp.StatusCode)
				}
			}
		}
	}
	return nil, errors.New("Error in answering query")
}

func Tts(answer interface{}) (map[string]interface{}, error) {
	client := &http.Client{}
	uri := "http://localhost:3003/tts"
	if jsontext, err := json.Marshal(answer); err == nil {
		if req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsontext)); err == nil {
			if rsp, err := client.Do(req); err == nil {
				if rsp.StatusCode == http.StatusOK {
					t := map[string]interface{}{}
					if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
						return t, nil
					} else {
						fmt.Println(rsp.StatusCode)
					}
				}
			}
		}
	}
	return nil, errors.New("Error in answering query")
}

func alexa(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if question, err := Stt(t); err == nil {
			if answer, err := Alpha(question); err == nil {
				if speechA, err := Tts(answer); err == nil {
					if audio, ok := speechA["speech"].(string); ok {
						u := map[string]interface{}{"speech": audio}
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(u)
					}
				}
			}
		}
	}
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alexa", alexa).Methods("POST")
	http.ListenAndServe(":3000", r)
}
