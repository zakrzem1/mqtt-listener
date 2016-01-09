package main_test

import (
  "testing"
  "github.com/stretchr/testify/assert"
	"github.com/zakrzem1/mqtt"
)

func TestDeserializeJson(t *testing.T) {
	reading, err := main.DeserializeJson([]byte("{\"temp\": 23.875, \"tstamp\": \"2015-12-19T12:41:42Z\"}"))
	assert.NotNil(t, reading)
	assert.True(t, reading.Temp > 0)

	assert.Nil(t, err)
//	123 34 116 101 109 112 34 58 32 50 52 46 54 56 55 44 32 34 116 115 116 97 109 112 34 58 32 34 50 48 49 53 45 49 50 45 50 51 84 48 54 58 53 57 58 53 52 90 34 125
//{"temp": "23.875", "tstamp": "2015-12-19T12:41:42Z"}

}