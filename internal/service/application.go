package service

import (
	"bytes"
	"client/internal/models"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (s *LevelSevenStressService) MakeRequestAndSend(request *models.Target) (uint16, int, error) {
	method := http.MethodGet
	var (
		body         io.Reader
		req          *http.Request
		err          error
		dumpWithBody bool
	)

	if request.Type.Uint16()&models.POST.Uint16() != 0 {
		method = http.MethodPost
		dumpWithBody = true
	} else if request.Type.Uint16()&models.PUT.Uint16() != 0 {
		method = http.MethodPut
		dumpWithBody = true
	} else if request.Type.Uint16()&models.DELETE.Uint16() != 0 {
		method = http.MethodDelete
	}

	if request.Type.Uint16()&models.POST.Uint16() != 0 || request.Type.Uint16()&models.PUT.Uint16() != 0 {
		if len(request.Payload) > 0 {
			if request.WithRandomizer {
				request.Payload = replace(request.Payload)
			}
			body = bytes.NewBuffer([]byte(request.Payload))
		}
	}

	scheme := "http://"
	if request.TargetPort == 443 || request.TargetPort == 8443 {
		scheme = "https://"
	}

	if request.TargetQuery != "" {
		if request.WithRandomizer {
			request.TargetQuery = replace(request.TargetQuery)
		}
	}

	if request.TargetIP != "" {
		req, err = http.NewRequest(method, scheme+request.TargetIP+request.TargetQuery, body)
	} else {
		req, err = http.NewRequest(method, scheme+request.TargetHost+request.TargetQuery, body)
	}
	if err != nil {
		return 0, 0, err
	}

	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	req.Host = request.TargetHost
	req.Header = make(http.Header)
	req.Header = request.Headers

	if request.Cookies != nil && len(request.Cookies) > 0 {
		for k, v := range request.Cookies {
			c := &http.Cookie{Name: k, Value: v}
			req.AddCookie(c)
		}
	}

	if request.WithRandomAgent {
		req.Header.Set("User-Agent", userAgents[ra.Intn(len(userAgents))])
	}

	bts, err := httputil.DumpRequest(s.request, dumpWithBody)
	if err != nil {
		return 0, 0, err
	}

	var client *http.Client

	if request.Type.Uint16()&models.HTTP1.Uint16() != 0 {
		client = s.clientHTTP1
	} else if request.Type.Uint16()&models.HTTP2.Uint16() != 0 {
		client = s.clientHTTP2
	} else if request.Type.Uint16()&models.HTTP3.Uint16() != 0 {
		client = s.clientHTTP3
	}

	r, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return 0, r.StatusCode, errors.New(strconv.Itoa(r.StatusCode))
	}

	return uint16(len(bts)), r.StatusCode, nil
}

func replace(s string) string {
	re := regexp.MustCompile(`\*\*__([^__]*)__\*\*`)
	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return s
	}
	return re.ReplaceAllString(s, generate(strings.TrimPrefix(strings.TrimSuffix(match[0], "_**"), "**_")))
}

func generate(typ string) string {
	var c = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w",
		"x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W",
		"X", "Y", "Z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".", "/", "?", "_"}

	var runes []string

	switch typ[:2] {
	case "AA":
		runes = c
	case "CC":
		runes = c[25:36]
	case "Cc":
		runes = c[:51]
	case "cc":
		runes = c[:26]
	case "NN":
		runes = c[50:61]
	}

	count, err := strconv.Atoi(typ[2:])
	if err != nil {
		return ""
	}

	var s strings.Builder

	for i := 0; i < count; i++ {
		s.WriteString(runes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(runes))])
	}

	return s.String()
}
