package cmq_go

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type CMQHttp struct {
	timeout     int
	isKeepAlive bool
	conn        *http.Client
}

func NewCMQHttp() *CMQHttp {
	return &CMQHttp{
		timeout:     10000,
		isKeepAlive: true,
		conn:        nil,
	}
}

var urlProxyMap sync.Map
var urlProxyMapLock sync.Mutex

var defaultProxyTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}


func (this *CMQHttp) getTransport(proxyUrlStr string) *http.Transport {
	if proxyUrlStr == "" {
		return defaultProxyTransport
	}
	var val interface{}
	var ok bool
	val, ok = urlProxyMap.Load(proxyUrlStr)
	if !ok {
		urlProxyMapLock.Lock()
		defer urlProxyMapLock.Unlock()
		val, ok = urlProxyMap.Load(proxyUrlStr)
		if !ok {
			proxyUrl, err := url.Parse(proxyUrlStr)
			if err != nil {
				panic(err)
			}
			proxyTransport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
			urlProxyMap.Store(proxyUrlStr, proxyTransport)
			val, _ = urlProxyMap.Load(proxyUrlStr)
		}
	}
	transport, ok := val.(*http.Transport)
	if !ok {
		panic("convert http transport err")
	}
	return transport
}

func (this *CMQHttp) request(method, urlStr, reqStr, proxyUrlStr string, userTimeout int) (result string, err error) {
	var client *http.Client
	timeout := 0
	if userTimeout >= 0 {
		timeout = userTimeout
	}
	//keepalive := 0
	//if this.isKeepAlive {
	//	keepalive = 30
	//}

	//if proxyUrlStr == "" {
	//	unproxyTransport := &http.Transport{
	//		Proxy: http.ProxyFromEnvironment,
	//		DialContext: (&net.Dialer{
	//			Timeout:   30 * time.Second,
	//			KeepAlive: time.Duration(keepalive) * time.Second,
	//			DualStack: true,
	//		}).DialContext,
	//		MaxIdleConns:          100,
	//		IdleConnTimeout:       90 * time.Second,
	//		TLSHandshakeTimeout:   10 * time.Second,
	//		ExpectContinueTimeout: 1 * time.Second,
	//	}
	//
	//	client = &http.Client{
	//		Timeout:   time.Duration(timeout) * time.Second,
	//		Transport: unproxyTransport,
	//	}
	//} else {
	//	proxyUrl, err := url.Parse(proxyUrlStr)
	//	if err != nil {
	//		panic(err)
	//	}
	//	proxyTransport := &http.Transport{
	//		Proxy: http.ProxyURL(proxyUrl),
	//		DialContext: (&net.Dialer{
	//			Timeout:   30 * time.Second,
	//			KeepAlive: time.Duration(keepalive) * time.Second,
	//			DualStack: true,
	//		}).DialContext,
	//		MaxIdleConns:          100,
	//		IdleConnTimeout:       90 * time.Second,
	//		TLSHandshakeTimeout:   10 * time.Second,
	//		ExpectContinueTimeout: 1 * time.Second,
	//	}
	//	client = &http.Client{
	//		Timeout:   time.Duration(timeout) * time.Second,
	//		Transport: proxyTransport,
	//	}
	//}

	client = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: this.getTransport(proxyUrlStr),
	}


	req, err := http.NewRequest(method, urlStr, bytes.NewReader([]byte(reqStr)))
	if err != nil {
		return "", fmt.Errorf("make http req error %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http error  %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read http resp body error %v", err)
	}
	result = string(body)
	return
}
