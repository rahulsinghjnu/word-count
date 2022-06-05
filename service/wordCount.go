package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type wcRequestBody struct {
	Message string `json:"message"`
}

func GetWordCount(url string, message []byte) (map[string]int, error) {
	requestBody := wcRequestBody{string(message)}

	jsonValue, err := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Resp Body Err: ", err)
	}
	wcResp := make(map[string]int)
	err = json.Unmarshal(body, &wcResp)
	return wcResp, nil
}
