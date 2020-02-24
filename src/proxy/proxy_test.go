package proxy

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"proxy/mockedHttpClient"
)

var (
	listener = "127.0.0.1:8080"
	backends = []Backend{
		Backend{"192.168.100.1:8080", "/urlinfo/1/"},
		Backend{"192.168.100.2:8080", "/urlinfo/2/"},
		Backend{"192.168.100.3:8080", "/urlinfo/3/"},
	}
)

func TestGetBackend(t *testing.T) {
	h := New(listener, backends, nil)
	be := h.getBackend("www.google.com")
	found := false
	for i := range backends {
		if backends[i].Endpoint == be.Endpoint && backends[i].Prefix == be.Prefix {
			found = true
		}
	}
	if !found {
		t.Fail()
	}
}

func TestAddPathToUrl(t *testing.T) {
	h := New(listener, backends, nil)
	if h.addPathToUrl("http://www.google.com", "search") != "http://www.google.com/search" {
		t.Fail()
	}
	if h.addPathToUrl("http://www.google.com/", "search") != "http://www.google.com/search" {
		t.Fail()
	}
	if h.addPathToUrl("http://www.google.com/", "/search") != "http://www.google.com/search" {
		t.Fail()
	}
}

func TestAddHostToXForwardHeader(t *testing.T) {
	h := New(listener, backends, nil)
	headers := http.Header{}
	h.addHostToXForwardHeader(headers, "www.google.com")
	assert.Equal(t, "www.google.com", headers.Get("X-Forwarded-For"))
	h.addHostToXForwardHeader(headers, "www.yahoo.com")
	assert.Equal(t, "www.google.com, www.yahoo.com", headers.Get("X-Forwarded-For"))
}

func TestCopyHeader(t *testing.T) {
	h := New(listener, backends, nil)
	src := http.Header{}
	dst := http.Header{}
	src.Add("Host", "127.0.0.1")
	h.copyHeader(dst, src)
	assert.Equal(t, "127.0.0.1", dst.Get("Host"))
}

func TestDelProxyHeaders(t *testing.T) {
	h := New(listener, backends, nil)
	headers := http.Header{}
	headers.Add("Connection", "close")
	h.delProxyHeaders(headers)
	assert.Equal(t, "", headers.Get("Connection"))
}

func TestServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cli := mockedhttp.NewMockHttpClient(ctrl)
	h := New(listener, backends, cli)

	req := httptest.NewRequest("GET", "http://127.0.0.1/"+"foo", nil)
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(nil, fmt.Errorf("Failed to GET foo."))
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	res := rw.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	rw = httptest.NewRecorder()
	ret := httptest.NewRecorder()
	http.Error(ret, "Failed to GET foo.", http.StatusBadRequest)
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(ret.Result(), nil)
	h.ServeHTTP(rw, req)
	res = rw.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	rw = httptest.NewRecorder()
	ret = httptest.NewRecorder()
	http.Error(ret, "Failed to GET foo.", http.StatusInternalServerError)
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(ret.Result(), nil)
	h.ServeHTTP(rw, req)
	res = rw.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	readAll = func(r io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("Failed to read from Reader.")
	}
	rw = httptest.NewRecorder()
	ret = httptest.NewRecorder()
	fmt.Fprintf(ret, "SAFE")
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(ret.Result(), nil)
	h.ServeHTTP(rw, req)
	res = rw.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	readAll = ioutil.ReadAll
	rw = httptest.NewRecorder()
	ret = httptest.NewRecorder()
	fmt.Fprintf(ret, "Unsafe")
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(ret.Result(), nil)
	h.ServeHTTP(rw, req)
	res = rw.Result()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	rw = httptest.NewRecorder()
	ret = httptest.NewRecorder()
	fmt.Fprintf(ret, "SAFE")
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(ret.Result(), nil)
	req1 := httptest.NewRequest("GET", "http://127.0.0.1/foo", nil)
	req1.Header.Add("X-Forwarded-For", "192.0.2.1")
	req1.RequestURI = ""
	cli.EXPECT().Do(req1).Return(nil, fmt.Errorf("Failed to DO HTTP request."))
	h.ServeHTTP(rw, req)
	res = rw.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	rw = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "http://127.0.0.1/"+"foo", nil)
	ret = httptest.NewRecorder()
	fmt.Fprintf(ret, "SAFE")
	cli.EXPECT().Get("http://192.168.100.1:8080/urlinfo/1/127.0.0.1/foo").Return(ret.Result(), nil)
	req1 = httptest.NewRequest("GET", "http://127.0.0.1/foo", nil)
	req1.Header.Add("X-Forwarded-For", "192.0.2.1")
	req1.RequestURI = ""
	cli.EXPECT().Do(req1).Return(ret.Result(), nil)
	h.ServeHTTP(rw, req)
	res = rw.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
