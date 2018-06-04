package translate

import (
	"encoding/json"
	"time"

	"github.com/wzshiming/requests"
)

var host = `https://translate.googleapis.com/translate_a/single`

var gt = requests.NewClient().
	SetLogLevel(requests.LogError).
	SetSkipVerify(true).
	NewRequest().
	SetTimeout(time.Second*2).
	SetMethod(requests.MethodPost).
	SetURL(host).
	SetUserAgent(ua).
	SetForm("client", "gtx").
	SetForm("dt", "t").
	SetForm("ie", "UTF-8").
	SetForm("oe", "UTF-8")

func GoogleTranslate(text, sourcelang, targetlang string) (string, error) {
	resp, err := gt.Clone().
		SetForm("q", text).
		SetForm("sl", sourcelang).
		SetForm("tl", targetlang).
		Do()
	if err != nil {
		return "", err
	}

	// [[["Hello there","你好",null,null,1]],null,"zh-CN",null,null,null,1,null,[["zh-CN"],null,[1],["zh-CN"]]]
	body := resp.Body()
	rms := []json.RawMessage{}
	for i := 0; i != 3; i++ {
		err := json.Unmarshal(body, &rms)
		if err != nil {
			return "", err
		}
		body = []byte(rms[0])
	}

	return string(body[1 : len(body)-1]), nil
}
