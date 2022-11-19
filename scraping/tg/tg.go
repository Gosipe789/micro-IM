package tg

import "usdt/scraping/req"

type TG struct {
	Url string
}

func (t *TG) SendMsg(parameter map[string]interface{}, body interface{}) error {
	// 对接api 发送post请求
	err := req.HttpPost(t.Url, parameter, body)
	if err != nil {
		return err
	}

	return nil
}
