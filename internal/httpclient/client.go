package httpclient

import (
	"net"
	"net/http"
	"time"
)

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,

		Transport: &http.Transport{
			// TCP 연결 설정
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,

			// TLS 핸드셰이크 타임아웃
			TLSHandshakeTimeout: 3 * time.Second,

			// 응답 헤더 수신 타임아웃
			ResponseHeaderTimeout: 5 * time.Second,

			// 커넥션 풀 설정
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     20,
			IdleConnTimeout:     90 * time.Second,

			// Expect: 100-continue 대기 시간
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
