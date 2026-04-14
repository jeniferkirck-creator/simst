package main

import (
	"client/config"
	"client/internal/models"
	"client/internal/service"
	"client/workers"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	cfgPath := flag.String("c", "", "Config file path")
	flag.Parse()

	configurationPath := *cfgPath

	if *cfgPath == "" {
		configurationPath, _ = os.Executable()
		configurationPath = filepath.Dir(configurationPath)
		configurationPath = filepath.Join(configurationPath, "config.json")
	}

	conf, err := config.Parse(configurationPath)
	if err != nil {
		panic(err)
	}

	ctx, chancel := context.WithCancel(context.Background())

	cfg, stresser, err := initWorkers(conf)
	if err != nil {
		panic(err)
	}

	go cfg.Start(ctx)
	go stresser.Start(ctx)
	fmt.Println("Starting app workers")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	<-signalChannel
	chancel()
	fmt.Println("Application shutdown")
	time.Sleep(30 * time.Second)
}

func initWorkers(cfg *config.Configuration) (*workers.ConfigurationLoader, *workers.StressWorker, error) {

	logChan := make(chan *models.Process)
	resChan := make(chan bool)
	configurationChannel := make(chan []*models.Target)

	var f *os.File
	var err error
	var serverService *service.ServerService

	if cfg.Server != nil {
		serverService = service.NewServerService(
			cfg.Server.RegisterLink,
			cfg.Server.TargetLink,
		)
		if err = serverService.Register(cfg.PublicIP); err != nil {
			return nil, nil, err
		}
	} else if cfg.TargetFilePatch != "" {
		f, err = os.Open(cfg.TargetFilePatch)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open targets file: %w", err)
		}
	}

	if err = service.NewLogger(logChan, resChan, cfg, f); err != nil {
		return nil, nil, err
	}

	http2client, err := service.NewHttp2Client()
	if err != nil {
		return nil, nil, err
	}

	httpStressService := service.NewLevelSevenStressService(
		service.NewHttp1Client(),
		http2client,
		service.NewHttp3Client(),
	)

	traStressService := service.NewLevelFourStressService(
		service.TcpSocketClient(),
		service.UdpSocketClient(),
	)

	configurationWorker := workers.NewConfigurationLoader(
		configurationChannel,
		cfg.TargetFilePatch,
		serverService,
	)

	stressWorker := workers.NewStressWorker(
		configurationChannel,
		logChan,
		httpStressService,
		traStressService,
		time.Duration(cfg.RequestTimeout)*time.Millisecond,
	)

	return configurationWorker, stressWorker, nil
}
