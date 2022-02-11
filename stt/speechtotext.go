package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func SpeechToText(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if encspeech, ok := t["speech"].(string); ok {
			fmt.Println(encspeech)
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
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func Service(speech []byte) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	check(err)

	req.Header.Set("Content-Type",
		"audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
	req.Header.Set("Accept", "json")

	rsp, err2 := client.Do(req)
	check(err2)

	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusOK {
		body, err3 := ioutil.ReadAll(rsp.Body)
		check(err3)
		return string(body), nil
	} else {
		return "", errors.New("cannot convert to speech to text")
	}
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
