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
			} else if rsp.StatusCode == http.StatusBadRequest {
				return nil, errors.New("400 error from Azure text to speech service: A required parameter is missing, empty, or null. Or, the value passed to either a required or optional parameter is invalid. A common reason is a header that's too long")
			} else if rsp.StatusCode == http.StatusUnauthorized {
				return nil, errors.New("401 error from Azure text to speech service: The request is not authorized. Make sure your subscription key or token is valid and in the correct region")
			} else if rsp.StatusCode == http.StatusTooManyRequests {
				return nil, errors.New("429 error from Azure text to speech service: You have exceeded the quota or rate of requests allowed for your subscription")
			} else if rsp.StatusCode == http.StatusBadGateway {
				return nil, errors.New("501 error from Azure text to speech service: There is a network or server-side problem. This status might also indicate invalid headers")
			} else {
				return nil, fmt.Errorf("error from Azure text to speech service, status code: %d", rsp.StatusCode)
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
