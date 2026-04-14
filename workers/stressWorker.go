package workers

import (
	"client/internal/models"
	"client/internal/service"
	"context"
	"time"
)

type out struct {
	pl   uint16
	code int
}

type StressWorker struct {
	targetChannel chan []*models.Target
	app, tra      service.Stresser
	logger        chan *models.Process
	requestDelay  time.Duration
}

func NewStressWorker(
	tch chan []*models.Target,
	logger chan *models.Process,
	app, tra service.Stresser,
	delay time.Duration) *StressWorker {
	return &StressWorker{
		requestDelay:  delay,
		targetChannel: tch,
		app:           app,
		tra:           tra,
		logger:        logger,
	}
}

func (s *StressWorker) Start(ctx context.Context) {
	loopCtx, cancelLoop := context.WithCancel(ctx)
	defer cancelLoop()

	logTicker := time.NewTicker(time.Minute)
	defer logTicker.Stop()

	outCH := make(chan out)
	load := make(chan struct{})

	var (
		dataLength   uint64
		codes        []int
		requestCount uint64
		loadedTests  int
	)

	go s.run(loopCtx, outCH, load)

	for {
		select {
		case <-load:
			loadedTests++
		case <-logTicker.C:
			s.logger <- &models.Process{
				Timestamp:     time.Now().Unix(),
				RequestsCount: requestCount,
				PayloadLength: dataLength,
				Codes:         codes,
			}
			dataLength = 0
			codes = make([]int, 0)
			requestCount = 0
		case <-ctx.Done():
			cancelLoop()
			for {
				if loadedTests == 0 {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			s.logger <- &models.Process{
				Timestamp:     time.Now().Unix(),
				RequestsCount: requestCount,
				PayloadLength: dataLength,
				Codes:         codes,
			}
			return
		case o := <-outCH:
			requestCount++
			codes = append(codes, o.code)
			dataLength += uint64(o.pl)
			loadedTests--
		}
	}
}

func (s *StressWorker) run(ctx context.Context, outputChan chan<- out, load chan struct{}) {
	loopTicker := time.NewTicker(s.requestDelay)
	defer loopTicker.Stop()

	var targets []*models.Target

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-s.targetChannel:
			targets = t
		case <-loopTicker.C:
			for i := range targets {
				load <- struct{}{}
				go s.work(targets[i], outputChan)
			}
		}
	}
}

func (s *StressWorker) work(t *models.Target, outputChan chan<- out) {
	var dataLength uint16
	var code int

	if t.Type.Uint16()&models.TCP.Uint16() != 0 {

		length, _, err := s.tra.MakeRequestAndSend(t)
		if err == nil {
			dataLength = length
		}
	} else if t.Type.Uint16()&models.UDP.Uint16() != 0 {
		length, _, err := s.tra.MakeRequestAndSend(t)
		if err == nil {
			dataLength = length
		}
	} else if t.Type.Uint16()&models.HTTP1.Uint16() != 0 ||
		t.Type.Uint16()&models.HTTP2.Uint16() != 0 ||
		t.Type.Uint16()&models.HTTP3.Uint16() != 0 {
		length, rc, err := s.app.MakeRequestAndSend(t)
		if err == nil {
			dataLength = length
			code = rc
		}
	}
	outputChan <- out{
		pl:   dataLength,
		code: code,
	}
}
