package requests

var defaul = NewClient().
	SetSkipVerify(true).
	WithLogger().
	WithCookieJar()

func NewRequest() *Request {
	return defaul.NewRequest()
}
