package service

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

func NewHttp1Client() *http.Client {
	return &http.Client{
		Transport: httpTransport(),
		Timeout:   15 * time.Second,
	}
}

func NewHttp2Client() (*http.Client, error) {
	transport := httpTransport()
	transport.TLSClientConfig.NextProtos = []string{"h2"}
	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}, nil
}

func NewHttp3Client() *http.Client {
	return &http.Client{
		Transport: &http3.Transport{
			QUICConfig: &quic.Config{
				MaxIdleTimeout:                 15 * time.Second,
				KeepAlivePeriod:                10 * time.Second,
				InitialStreamReceiveWindow:     6 * 1024 * 1024,
				InitialConnectionReceiveWindow: 15 * 1024 * 1024,
			},

			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
				MaxVersion:         tls.VersionTLS13,
			},
		},
		Timeout: 15 * time.Second,
	}
}

func httpTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     1000,
		IdleConnTimeout:     20 * time.Second,

		DisableCompression: true,
		DisableKeepAlives:  false,

		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 0,

		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
			MaxVersion:         tls.VersionTLS13,
			NextProtos:         []string{"h3"},
		},
	}
}
