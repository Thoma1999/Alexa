package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
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

//Create xml request for tts serices
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
	jsonData := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err == nil {
		if words, ok := jsonData["text"].(string); ok {
			if myXml, err := GenerateXMLByte(words); err == nil {
				if speech, code, err := Service([]byte(xml.Header + string(myXml))); err == nil {
					//b64 encode speech
					encSpeech := b64.StdEncoding.EncodeToString(speech)
					jsonResponse := map[string]interface{}{"speech": encSpeech}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(jsonResponse)
				} else {
					//Output error from Azure tts service
					http.Error(w, err.Error(), code)
				}
			} else {
				http.Error(w, "Error generating xml for request", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Error getting text field from JSON", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
	}
}

func Service(text []byte) ([]byte, int, error) {
	client := &http.Client{}
	if req, err := http.NewRequest("POST", URI, bytes.NewBuffer(text)); err == nil {
		req.Header.Set("Content-Type", "application/ssml+xml")
		req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
		req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
		if rsp, err := client.Do(req); err == nil {
			if rsp.StatusCode == http.StatusOK {
				if body, err := ioutil.ReadAll(rsp.Body); err == nil {
					return body, http.StatusOK, nil
				}
				//Return error message for each status code
			} else if rsp.StatusCode == http.StatusBadRequest {
				return nil, http.StatusBadRequest, errors.New("error from Azure text to speech service: A required parameter is missing, empty, or null. Or, the value passed to either a required or optional parameter is invalid. A common reason is a header that's too long")
			} else if rsp.StatusCode == http.StatusUnauthorized {
				return nil, http.StatusUnauthorized, errors.New("error from Azure text to speech service: The request is not authorized. Make sure your subscription key or token is valid and in the correct region")
			} else if rsp.StatusCode == http.StatusTooManyRequests {
				return nil, http.StatusTooManyRequests, errors.New("error from Azure text to speech service: You have exceeded the quota or rate of requests allowed for your subscription")
			} else if rsp.StatusCode == http.StatusBadGateway {
				return nil, http.StatusBadGateway, errors.New("error from Azure text to speech service: There is a network or server-side problem. This status might also indicate invalid headers")
			} else {
				return nil, rsp.StatusCode, errors.New("error from Azure text to speech service")
			}
		}
	}
	return nil, http.StatusInternalServerError, errors.New("error making request to Azure text to speech service")
}

func main() {
	//Create Router and listen for POST requests on localhost:3003/tts
	r := mux.NewRouter()
	r.HandleFunc("/tts", TextToSpeech).Methods("POST")
	http.ListenAndServe(":3003", r)
}
