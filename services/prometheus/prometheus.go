package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Service struct {
	address string
	logger  *logrus.Logger

	blocksSigned          prometheus.Counter
	blocksMissedTotal     prometheus.Counter
	blocksMissedCurrent   prometheus.Gauge
	missedBlocksThreshold prometheus.Counter
	sleep                 prometheus.Counter
}

func New(address string, logger *logrus.Logger) (*Service, error) {
	svc := &Service{
		address: address,
		logger:  logger,
	}

	svc.missedBlocksThreshold = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minter_sentinel_missed_blocks_threshold",
		Help: "Missed blocks threshold before masternode will go off",
	})

	svc.sleep = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minter_sentinel_sleep",
		Help: "Number of seconds to sleep between checking for missed blocks",
	})

	svc.blocksSigned = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minter_sentinel_blocks_signed",
		Help: "The total number of signed blocks",
	})

	svc.blocksMissedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minter_sentinel_blocks_missed_total",
		Help: "The total number of missed blocks",
	})

	svc.blocksMissedCurrent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "minter_sentinel_blocks_missed_current",
		Help: "The current number of missed blocks",
	})

	return svc, nil
}

func (s *Service) Start() error {
	http.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(s.address, nil)
}

func (s *Service) SetMissedBlocksThreshold(value int) {
	s.missedBlocksThreshold.Add(float64(value))
}

func (s *Service) SetSleep(value int) {
	s.sleep.Add(float64(value))
}

func (s *Service) BlocksSignedIncrement() {
	s.blocksSigned.Inc()
}

func (s *Service) BlocksMissedIncrement() {
	s.blocksMissedTotal.Inc()
}

func (s *Service) SetBlocksMissedCurrent(value int) {
	s.blocksMissedCurrent.Set(float64(value))
}
