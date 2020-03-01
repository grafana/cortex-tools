package reader

import (
	"context"
	"fmt"
	"sync"

	cortex_chunk "github.com/cortexproject/cortex/pkg/chunk"
	cortex_storage "github.com/cortexproject/cortex/pkg/chunk/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/grafana/cortextool/pkg/chunk"
	"github.com/grafana/cortextool/pkg/chunk/storage"
)

var (
	SentChunks = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cortex",
		Name:      "reader_sent_chunks_total",
		Help:      "The total number of chunks sent by this reader.",
	})
)

// Config is a config for a Reader
type Config struct {
	StorageType   string                `yaml:"storage_type"`
	StorageConfig cortex_storage.Config `yaml:"storage"`
	NumWorkers    int                   `yaml:"num_workers"`
}

// Reader collects and forwards chunks according to it's planner
type Reader struct {
	cfg Config
	id  string // ID is the configured as the reading prefix and the shards assigned to the reader

	scanner          chunk.Scanner
	planner          *Planner
	workerGroup      sync.WaitGroup
	scanRequestsChan chan chunk.ScanRequest
	err              error
	quit             chan struct{}
}

// NewReader returns a Reader struct
func NewReader(cfg Config, plannerCfg PlannerConfig) (*Reader, error) {
	planner, err := NewPlanner(plannerCfg)
	if err != nil {
		return nil, err
	}

	scanner, err := storage.NewChunkScanner(cfg.StorageType, cfg.StorageConfig)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%d_%d", plannerCfg.FirstShard, plannerCfg.LastShard)

	// Default to one worker if none is set
	if cfg.NumWorkers < 1 {
		cfg.NumWorkers = 1
	}

	return &Reader{
		cfg:              cfg,
		id:               id,
		planner:          planner,
		scanner:          scanner,
		scanRequestsChan: make(chan chunk.ScanRequest),
		quit:             make(chan struct{}),
	}, nil
}

// Run initializes the writer workers
func (r *Reader) Run(ctx context.Context, outChan chan cortex_chunk.Chunk) {
	defer close(outChan)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < r.cfg.NumWorkers; i++ {
		g.Go(func() error {
			return r.readLoop(ctx, outChan)
		})
	}

	scanRequests := r.planner.Plan()
	logrus.Infof("built %d plans for reading", len(scanRequests))

	// feeding scan requests to workers
	for _, req := range scanRequests {
		select {
		case <-ctx.Done():
			logrus.Info("shutting down reader because context was cancelled")
			return
		case r.scanRequestsChan <- req:
			continue
		case <-r.quit:
			return
		}
	}

	// all scan requests are fed, close the channel
	close(r.scanRequestsChan)

	r.err = g.Wait()
}

func (r *Reader) readLoop(ctx context.Context, outChan chan cortex_chunk.Chunk) error {
	defer r.workerGroup.Done()

	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("shutting down reader because context was cancelled")
			return nil
		case req, open := <-r.scanRequestsChan:
			if !open {
				return nil
			}

			logEntry := logrus.WithFields(logrus.Fields{
				"table": req.Table,
				"user":  req.User,
				"shard": req.Prefix})

			logEntry.Infoln("attempting  scan request")
			err := r.scanner.Scan(ctx, req, func(i cortex_chunk.Chunk) bool {
				// while this does not mean chunk is sent by scanner, this is the closest we can get
				SentChunks.Inc()
				return true
			}, outChan)

			if err != nil {
				logEntry.WithError(err).Errorln("error scanning chunks")
				return fmt.Errorf("scan request failed, %v", req)
			}

			logEntry.Infoln("completed scan request")
		}
	}
}

func (r *Reader) Stop() {
	close(r.quit)
}

func (r *Reader) Err() error {
	return r.err
}
