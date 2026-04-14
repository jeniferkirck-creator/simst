package service

import (
	"bytes"
	"client/config"
	"client/internal/models"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewLogger(ch chan *models.Process, response chan bool, cfg *config.Configuration, f *os.File) error {
	if f == nil && (cfg.Server != nil && cfg.Server.ReportLink == "") {
		return errors.New("must specify a file to log or host to send")
	}

	reportPath := ""
	if cfg.Server != nil && cfg.Server.ReportLink != "" {
		reportPath = cfg.Server.ReportLink
	}

	go process(ch, response, reportPath, f)

	return nil
}

func process(ch chan *models.Process, response chan bool, serverPath string, f *os.File) {
	for msg := range ch {
		now := time.Now()
		msg.Timestamp = now.Unix()
		if f != nil {
			var codes []string
			for i := range msg.Codes {
				codes = append(codes, strconv.Itoa(msg.Codes[i]))
			}
			s := fmt.Sprintf("%s [ requests = %d]\n\t[ %s ]\n",
				now.Format(time.RFC3339),
				msg.RequestsCount,
				strings.Join(codes, ", "))
			writeIntoLogfile(f, s, response)
			printMessage(s)
		} else {
			sendProcess(msg, response, serverPath)
		}
	}
}

func writeIntoLogfile(f *os.File, s string, response chan bool) {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	if _, err := f.WriteString(s); err != nil {
		response <- false
	}
	response <- true
}

func printMessage(msg string) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Println(msg)
}

func sendProcess(p *models.Process, response chan bool, serverPath string) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	bts, err := json.Marshal(p)
	if err != nil {
		response <- false
		return
	}
	resp, err := client.Post(serverPath, "Content-Type: application/json", bytes.NewBuffer(bts))
	if err != nil {
		response <- false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response <- false
	}
	response <- true
}
