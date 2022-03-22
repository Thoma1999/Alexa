package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".stt.speech.microsoft.com/" +
		"speech/recognition/conversation/cognitiveservices/v1?" +
		"language=en-US"
	KEY = "d76745e51adf4408b1f29d7a4362dc39"
	MSG = "Sorry i did not understand that"
)

func SpeechToText(w http.ResponseWriter, r *http.Request) {
	jsonData := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err == nil {
		if encSpeech, ok := jsonData["speech"].(string); ok {
			if speech, err := b64.StdEncoding.DecodeString(encSpeech); err == nil {
				if words, err := Service(speech); err == nil {
					jsonResponse := map[string]interface{}{"text": words}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(jsonResponse)
				} else {
					//Output error from Azure stt service
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
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

func Service(speech []byte) (string, error) {
	client := &http.Client{}
	if req, err := http.NewRequest("POST", URI, bytes.NewReader(speech)); err == nil {
		req.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
		req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
		req.Header.Set("Accept", "json")
		if rsp, err := client.Do(req); err == nil {
			if rsp.StatusCode == http.StatusOK {
				jsonResponse := map[string]interface{}{}
				if err := json.NewDecoder(rsp.Body).Decode(&jsonResponse); err == nil {
					if jsonResponse["RecognitionStatus"].(string) == "Success" {
						return jsonResponse["DisplayText"].(string), nil
					} else {
						// Returns a misunderstanding message if no speech could be detected in the audio
						fmt.Println(jsonResponse["RecognitionStatus"].(string))
						return MSG, nil
					}
				}
			} else if rsp.StatusCode == http.StatusBadRequest {
				// Returns a misunderstanding message for incrorrect language or invalid audio file
				fmt.Println("400 error from Azure speech to text service. The language code wasn't provided, the language isn't supported, or the audio file is invalid")
				return MSG, nil
				//Return error message for each status code
			} else if rsp.StatusCode == http.StatusRequestTimeout {
				return "", errors.New("408 error from Azure speech to text service. The error most likely occurs because no audio data is being sent to the service. This error also might be caused by network issues")
			} else if rsp.StatusCode == http.StatusUnauthorized {
				return "", errors.New("401 error from Azure speech to text service. A subscription key or an authorization token is invalid in the specified region, or an endpoint is invalid")
			} else if rsp.StatusCode == http.StatusForbidden {
				return "", errors.New("403 error from Azure speech to text service. A subscription key or authorization token is missing")
			} else {
				return "", fmt.Errorf("error from Azure speech to text service, status code: %d", rsp.StatusCode)
			}
		}
	}
	return "", errors.New("error making requst to Azure speech to text service")
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/stt", SpeechToText).Methods("POST")
	http.ListenAndServe(":3002", r)
}
