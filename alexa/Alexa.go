package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Makes POST request to given microservice
func Service(jsonData interface{}, uri string, serviceName string) (map[string]interface{}, error) {
	client := &http.Client{}
	if jsontext, err := json.Marshal(jsonData); err == nil {
		if req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsontext)); err == nil {
			if rsp, err := client.Do(req); err == nil {
				if rsp.StatusCode == http.StatusOK {
					jsonResponse := map[string]interface{}{}
					if err := json.NewDecoder(rsp.Body).Decode(&jsonResponse); err == nil {
						return jsonResponse, nil
					}
				} else {
					// Returns and error message and statuscode from the microservice called
					return nil, fmt.Errorf("%d error from "+serviceName+" service", rsp.StatusCode)
				}
			}
		}
	}
	return nil, errors.New("error making request to " + serviceName + " service")
}

func alexa(w http.ResponseWriter, r *http.Request) {
	jsonData := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err == nil {
		//Get text from stt
		if question, err := Service(jsonData, "http://localhost:3002/stt", "Speech to text"); err == nil {
			//Get answer to query from alpha
			if answer, err := Service(question, "http://localhost:3001/alpha", "Wolfram Alpha"); err == nil {
				//Get spoken answer from tts
				if speechOutput, err := Service(answer, "http://localhost:3003/tts", "Text to speech"); err == nil {
					//Get encoded speech as string from speechOutput json
					if audio, ok := speechOutput["speech"].(string); ok {
						jsonResponse := map[string]interface{}{"speech": audio}
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(jsonResponse)
					}
				} else {
					//Output error from tts
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				//Output error from alpha
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			//Output error from stt
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		//Output error if json is missing from request
		fmt.Println("unable to get json data from request")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	//Create Router and listen for POST requests on localhost:3000/alexa
	r := mux.NewRouter()
	r.HandleFunc("/alexa", alexa).Methods("POST")
	http.ListenAndServe(":3000", r)
}
