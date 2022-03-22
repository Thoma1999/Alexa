package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

// Makes POST request to given microservice
func Service(jsonData interface{}, uri string, serviceName string) (map[string]interface{}, int, error) {
	client := &http.Client{}
	if jsontext, err := json.Marshal(jsonData); err == nil {
		if req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsontext)); err == nil {
			if rsp, err := client.Do(req); err == nil {
				if rsp.StatusCode == http.StatusOK {
					jsonResponse := map[string]interface{}{}
					if err := json.NewDecoder(rsp.Body).Decode(&jsonResponse); err == nil {
						return jsonResponse, http.StatusOK, nil
					}
				} else {
					// Returns and error message and statuscode from the microservice called
					return nil, rsp.StatusCode, errors.New("error from " + serviceName + " service")
				}
			}
		} else {
			return nil, http.StatusBadRequest, errors.New("error making request to " + serviceName + " service")
		}
	}
	return nil, http.StatusBadRequest, errors.New("error marshalling JSON for" + serviceName + " service")
}

func alexa(w http.ResponseWriter, r *http.Request) {
	jsonData := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err == nil {
		//Get text from stt
		if question, code, err := Service(jsonData, "http://localhost:3002/stt", "Speech to text"); err == nil {
			//Get answer to query from alpha
			if answer, code, err := Service(question, "http://localhost:3001/alpha", "Alpha"); err == nil {
				//Get spoken answer from tts
				if speechOutput, code, err := Service(answer, "http://localhost:3003/tts", "Text to speech"); err == nil {
					//Get encoded speech as string from speechOutput json
					if audio, ok := speechOutput["speech"].(string); ok {
						jsonResponse := map[string]interface{}{"speech": audio}
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(jsonResponse)
					}
				} else {
					//Output error from tts
					http.Error(w, err.Error(), code)
				}
			} else {
				//Output error from alpha
				http.Error(w, err.Error(), code)
			}
		} else {
			//Output error from stt
			http.Error(w, err.Error(), code)
		}
	} else {
		//Output error if json is missing from request
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
	}
}

func main() {
	//Create Router and listen for POST requests on localhost:3000/alexa
	r := mux.NewRouter()
	r.HandleFunc("/alexa", alexa).Methods("POST")
	http.ListenAndServe(":3000", r)
}
