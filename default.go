package requests

var defaul = NewClient()

func NewRequest() *Request {
	return defaul.NewRequest()
}
