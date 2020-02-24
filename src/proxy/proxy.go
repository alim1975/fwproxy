package proxy

import (
	"github.com/sirupsen/logrus"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

var (
	logger  = logrus.StandardLogger()
	readAll = ioutil.ReadAll
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (resp *http.Response, err error)
}

type Config struct {
	Listener string    `yaml:"listener"`
	Urldb    []Backend `yaml:"urldb"`
}

type Backend struct {
	Endpoint string `yaml:"endpoint"`
	Prefix   string `yaml:"prefix"`
}

type HttpProxyWithFirewall struct {
	listener string
	db       []Backend
	client   HttpClient
}

func New(listener string, dbase []Backend, cli HttpClient) *HttpProxyWithFirewall {
	return &HttpProxyWithFirewall{listener: listener, db: dbase, client: cli}
}

func (h *HttpProxyWithFirewall) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	logger.Debug("Receive request: ", req.Method, " Remote Address: ", req.RemoteAddr, " URL: ", req.URL)
	url := h.addPathToUrl(req.URL.Host, req.URL.EscapedPath())
	be := h.getBackend(url)
	URL := h.addPathToUrl(be.Endpoint+be.Prefix, url)
	resp, err := h.client.Get("http://" + strings.TrimSuffix(URL, "/"))
	if err != nil {
		logger.Error("URL database lookup error: ", err)
		http.Error(res, "Failed to lookup URL database.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		logger.Error("URL database lookup returned HTTP Bad Request: ", resp)
		http.Error(res, "URL database lookup failed.", http.StatusInternalServerError)
		return
	} else if resp.StatusCode == http.StatusInternalServerError {
		logger.Error("URL database lookup returned Internal Sever Error", resp)
		http.Error(res, "URL database lookup failed.", http.StatusInternalServerError)
		return
	}
	text, err := readAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read HTTP response body.", err)
		http.Error(res, "URL database lookup failed.", http.StatusInternalServerError)
		return
	}
	if string(text) == "SAFE" {
		req.RequestURI = ""
		h.delProxyHeaders(req.Header)
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			h.addHostToXForwardHeader(req.Header, clientIP)
		}
		logger.Debug("DO(req)", req)
		resp, err = h.client.Do(req)
		if err != nil {
			logger.Error("ServeHTTP error: ", err)
			http.Error(res, "Failed to DO HTTP request.", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		logger.Debug(req.RemoteAddr, " ", resp.Status)
		h.delProxyHeaders(resp.Header)
		h.copyHeader(res.Header(), resp.Header)
		res.WriteHeader(resp.StatusCode)
		io.Copy(res, resp.Body)
	} else {
		logger.Warn("Dropping client request because URL databse says: ", string(text))
		http.Error(res, "Firewall blocked the request.", http.StatusForbidden)
	}
}

func (h *HttpProxyWithFirewall) Start() error {
	logger.Info("Starting HTTP Proxy With Firewall @ ", h.listener)
	return http.ListenAndServe(h.listener, h)
}

func (h *HttpProxyWithFirewall) getBackend(url string) Backend {
	hash := fnv.New32a()
	hash.Write([]byte(url))
	return h.db[hash.Sum32()%uint32(len(h.db))]
}

func (h *HttpProxyWithFirewall) addPathToUrl(url, path string) string {
	urlEndsWithSlash := strings.HasSuffix(url, "/")
	pathBeginsWithSlash := strings.HasPrefix(path, "/")

	switch {
	case urlEndsWithSlash && pathBeginsWithSlash:
		return url + path[1:]
	case !urlEndsWithSlash && !pathBeginsWithSlash:
		return url + "/" + path
	}

	return url + path
}

func (h *HttpProxyWithFirewall) addHostToXForwardHeader(header http.Header, host string) {
	if prev, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prev, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

var proxyHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func (h *HttpProxyWithFirewall) copyHeader(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func (h *HttpProxyWithFirewall) delProxyHeaders(header http.Header) {
	for _, h := range proxyHeaders {
		header.Del(h)
	}
}
