package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Configuration struct {
	TargetFilePatch string `json:"target_file_patch,omitempty"`
	TestDuration    uint8  `json:"test_duration"`   // in minutes
	RequestTimeout  uint8  `json:"request_timeout"` // in microseconds
	SteppingPayload struct {
		StepDuration        uint8 `json:"step_duration"`         // in minutes
		StartRequestTimeout uint8 `json:"start_request_timeout"` // in microseconds
		Step                uint8 `json:"step"`                  // percents
	} `json:"stepping_payload,omitempty"`
	Server *struct {
		TargetLink   string `json:"target_link"`
		ReportLink   string `json:"report_link"`
		RegisterLink string `json:"register_link"`
	} `json:"server,omitempty"`
	ResultFilePatch string `json:"result_file_patch"`
	PublicIP        string `json:"public_ip,omitempty"`
}

func Parse(fp string) (*Configuration, error) {
	f, err := os.Open(fp)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if os.IsNotExist(err) {
		return defaultConfiguration(), nil
	}
	defer f.Close()

	var config Configuration
	if err = json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func defaultConfiguration() *Configuration {
	ex, _ := os.Executable()
	ex = filepath.Dir(ex)
	return &Configuration{
		TargetFilePatch: filepath.Join(ex, "targets.json"),
		TestDuration:    30,
		RequestTimeout:  30,
		ResultFilePatch: filepath.Join(ex, fmt.Sprintf("%s.txt", time.Now().Format(time.DateOnly))),
	}
}
