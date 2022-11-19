package req

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Http struct {
}

// Get get获取数据
func (h *Http) Get(url string, data interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	//resp, err := http.Get(url)
	// 设置接口超时时间
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &data)
}

func HttpPost(url string, parameter map[string]interface{}, data interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	// parameter 转为json字符串
	parameterJson, err := json.Marshal(parameter)
	if err != nil {
		return err
	}
	payload := strings.NewReader(string(parameterJson))
	response, err := client.Post(url, "application/json", payload)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return json.Unmarshal(body, &data)
}
