package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type Speak struct {
	XMLName xml.Name `xml:"speak"`
	Version string   `xml:"version,attr"`
	Lang    string   `xml:"xml:lang,attr"`
	Voice   struct {
		Text string `xml:",chardata"`
		Lang string `xml:"xml:lang,attr"`
		Name string `xml:"name,attr"`
	} `xml:"voice"`
}

const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".tts.speech.microsoft.com/" +
		"cognitiveservices/v1"
	KEY      = "d76745e51adf4408b1f29d7a4362dc39"
	VERSION  = "1.0"
	LANGUAGE = "en-US"
	NAME     = "en-US-JennyNeural"
)

func GenerateXMLByte(text string) ([]byte, error) {
	req := &Speak{}
	req.Version = VERSION
	req.Lang = LANGUAGE
	req.Voice.Text = text
	req.Voice.Lang = LANGUAGE
	req.Voice.Name = NAME
	if u, err := xml.Marshal(req); err == nil {
		return u, nil
	} else {
		return nil, err
	}
}

func TextToSpeech(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if words, ok := t["text"].(string); ok {
			if myXml, err := GenerateXMLByte(words); err == nil {
				if speech, err := Service([]byte(xml.Header + string(myXml))); err == nil {
					sEnc := b64.StdEncoding.EncodeToString(speech)
					u := map[string]interface{}{"speech": sEnc}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(u)
				} else {
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

func Service(text []byte) ([]byte, error) {
	client := &http.Client{}
	if req, err := http.NewRequest("POST", URI, bytes.NewBuffer(text)); err == nil {
		req.Header.Set("Content-Type", "application/ssml+xml")
		req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
		req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
		if rsp, err := client.Do(req); err == nil {
			if rsp.StatusCode == http.StatusOK {
				if body, err := ioutil.ReadAll(rsp.Body); err == nil {
					return body, nil
				}
			} else {
				return nil, errors.New("error from Azure text to speech service, status code: " + string(rsp.StatusCode))
			}
		}
	}
	return nil, errors.New("error making request to Azure text to speech service")
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/tts", TextToSpeech).Methods("POST")
	http.ListenAndServe(":3003", r)
}
