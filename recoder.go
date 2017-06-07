package httpmock

import "net/http/httptest"

type CloseNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func NewCloseNotifyingRecorder() *CloseNotifyingRecorder {
	return &CloseNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *CloseNotifyingRecorder) Close() {
	c.closed <- true
}

func (c *CloseNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}
