#!/bin/sh
echo "{\"speech\":\"`base64 -i question.wav`\"}" > input
JSON2=`
`
echo $JSON2 | cut -d '"' -f4 | base64 -d > answer.wav

curl -v -X POST -d "{\"speech\":\"`base64 -i question.wav`\"}" localhost:3000/alexa