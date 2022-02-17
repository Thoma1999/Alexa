package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func GetResponse(req *http.Request) (string, error) {
	client := &http.Client{}
	if rsp, err2 := client.Do(req); err2 == nil {
		defer rsp.Body.Close()
		if rsp.StatusCode == http.StatusOK {
			if body, err3 := ioutil.ReadAll(rsp.Body); err3 == nil {
				return string(body), nil
			} else {
				return "", err3
			}
		} else {
			return "", errors.New("request was not understood")
		}
	} else {
		return "", err2
	}
}

func alexa(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if encspeech, ok := t["speech"].(string); ok {
			if req, err := http.NewRequest("POST", "localhost:3002/stt", bytes.NewReader([]byte(encspeech))); err == nil {
				if question, err := GetResponse(req); err == nil {
					if req, err := http.NewRequest("POST", "localhost:3001/alpha", bytes.NewReader([]byte(question))); err == nil {
						if answer, err := GetResponse(req); err == nil {
							if encanswer, err := http.NewRequest("POST", "localhost:3003/tts", bytes.NewReader([]byte(answer))); err == nil {
								u := map[string]interface{}{"speech": encanswer}
								w.WriteHeader(http.StatusOK)
								json.NewEncoder(w).Encode(u)
								if speech, err := b64.StdEncoding.DecodeString(encspeech); err == nil {
									err = ioutil.WriteFile("answer.wav", speech, 0644)
									check(err)
								} else {
									w.WriteHeader(http.StatusBadRequest)
								}
							} else {
								w.WriteHeader(http.StatusBadRequest)
							}
						} else {
							w.WriteHeader(http.StatusBadRequest)
						}
					} else {
						w.WriteHeader(http.StatusBadRequest)
					}
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alexa", alexa).Methods("POST")
	http.ListenAndServe(":3000", r)
}
