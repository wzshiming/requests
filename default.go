package requests

var defaul = NewClient().
	SetSkipVerify(true).
	WithLogger().
	SetLogLevel(LogMessageHead).
	WithCookieJar()

func NewRequest() *Request {
	return defaul.NewRequest()
}
