package main
import (
  "bytes"
  "errors"
  "io/ioutil"
  "net/http"
)

const (
  REGION = "uksouth"
  URI    = "https://" + REGION + ".tts.speech.microsoft.com/" +
           "cognitiveservices/v1"
  KEY    = "19c1cb3c0aa848608fed5a5a8a23d640"
)

func check( e error ) { if e != nil { panic( e ) } }

func TextToSpeech( text []byte ) ( []byte, error ) {
  client   := &http.Client{}
  req, err := http.NewRequest( "POST", URI, bytes.NewBuffer( text ) )
  check( err )

  req.Header.Set( "Content-Type", "application/ssml+xml" )
  req.Header.Set( "Ocp-Apim-Subscription-Key", KEY )
  req.Header.Set( "X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm" )

  rsp, err2 := client.Do( req )
  check( err2 )

  defer rsp.Body.Close()

  if rsp.StatusCode == http.StatusOK {
    body, err3 := ioutil.ReadAll( rsp.Body )
    check( err3 )
    return body,  nil
  } else {
    return nil, errors.New( "cannot convert text to speech" )
  }
} 

func main() {
  text, err := ioutil.ReadFile( "text.xml" )
  check( err )
  speech, err2 := TextToSpeech( text )
  check( err2 )
  err3 := ioutil.WriteFile( "speech.wav", speech, 0644 )
  check( err3 )
}

