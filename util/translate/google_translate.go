package translate

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/wzshiming/requests"
)

var (
	host = `https://translate.googleapis.com/translate_a/single`

	cli = requests.NewClient().
		SetLogLevel(requests.LogError).
		SetSkipVerify(true)

	gt = cli.NewRequest().
		SetTimeout(time.Second*2).
		SetMethod(requests.MethodPost).
		SetURLByStr(host).
		SetUserAgent(ua).
		SetForm("client", "gtx").
		SetForm("dt", "t").
		SetForm("ie", "UTF-8").
		SetForm("oe", "UTF-8")
)

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

	ret := messages{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return "", err
	}
	return ret.String(), nil
}

type messages []string

func (m *messages) String() string {
	text := strings.Join([]string(*m), " ")
	text = strings.Replace(text, "\\n", "\n", -1)
	return text
}

func (m *messages) UnmarshalJSON(body []byte) error {
	rms := []json.RawMessage{}
	for i := 0; i != 2; i++ {
		err := json.Unmarshal(body, &rms)
		if err != nil {
			return err
		}
		body = []byte(rms[0])
	}
	for _, v := range rms {
		b := []json.RawMessage{}
		err := json.Unmarshal([]byte(v), &b)
		if err != nil {
			return err
		}
		by := b[0]
		by = by[1 : len(by)-1]
		*m = append(*m, string(by))
	}
	return nil
}
