package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Service(jsonData interface{}, uri string, serviceName string) (map[string]interface{}, error) {
	client := &http.Client{}
	if jsontext, err := json.Marshal(jsonData); err == nil {
		if req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsontext)); err == nil {
			if rsp, err := client.Do(req); err == nil {
				if rsp.StatusCode == http.StatusOK {
					t := map[string]interface{}{}
					if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
						return t, nil
					}
				}
			}
		}
	}
	return nil, errors.New("Error in " + serviceName + " service")
}

func alexa(w http.ResponseWriter, r *http.Request) {
	speechInput := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&speechInput); err == nil {
		if question, err := Service(speechInput, "http://localhost:3002/stt", "Speech to text"); err == nil {
			if answer, err := Service(question, "http://localhost:3001/alpha", "Wolfram Alpha"); err == nil {
				if speechOutput, err := Service(answer, "http://localhost:3003/tts", "Text to speech"); err == nil {
					if audio, ok := speechOutput["speech"].(string); ok {
						u := map[string]interface{}{"speech": audio}
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(u)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
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
