package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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

func SpeechToText(speech []byte) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	check(err)

	req.Header.Set("Content-Type",
		"audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)

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
	speech, err1 := ioutil.ReadFile("speech.wav")
	check(err1)
	text, err2 := SpeechToText(speech)
	check(err2)
	fmt.Println(text)
}
