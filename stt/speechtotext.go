package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".stt.speech.microsoft.com/" +
		"speech/recognition/conversation/cognitiveservices/v1?" +
		"language=en-US"
	KEY = "d76745e51adf4408b1f29d7a4362dc39"
)

func SpeechToText(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if encspeech, ok := t["speech"].(string); ok {
			if speech, err := b64.StdEncoding.DecodeString(encspeech); err == nil {
				if words, err := Service(speech); err == nil {
					u := map[string]interface{}{"text": words}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(u)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusBadGateway)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
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
				t := map[string]interface{}{}
				if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
					return t["DisplayText"].(string), nil
				}
			}
		}
	}
	return "", errors.New("cannot convert to speech to text")

}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/stt", SpeechToText).Methods("POST")
	http.ListenAndServe(":3002", r)
}

/*
func main() {
	speech, err1 := ioutil.ReadFile("speech.wav")
	check(err1)
	text, err2 := SpeechToText(speech)
	check(err2)
	fmt.Println(text)
}
*/
