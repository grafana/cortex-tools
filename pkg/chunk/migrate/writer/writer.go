package writer

import (
	"context"
	"time"

	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/cortexproject/cortex/pkg/chunk/storage"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/validation"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	ReceivedChunks = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex",
		Name:      "migration_writer_received_chunks_total",
		Help:      "The total number of chunks received by this writer",
	}, nil)
)

// Config configures the Writer struct
type Config struct {
	StorageConfig storage.Config     `yaml:"storage"`
	SchemaConfig  chunk.SchemaConfig `yaml:"schema"`
}

// Writer receives chunks and stores them in a storage backend
type Writer struct {
	cfg        Config
	chunkStore chunk.Store
	mapper     Mapper

	quit chan struct{}
}

// NewWriter returns a Writer object
func NewWriter(cfg Config, mapper Mapper) (*Writer, error) {
	overrides, err := validation.NewOverrides(validation.Limits{})
	if err != nil {
		return nil, err
	}

	chunkStore, err := storage.NewStore(cfg.StorageConfig, chunk.StoreConfig{}, cfg.SchemaConfig, overrides)
	if err != nil {
		return nil, err
	}

	writer := Writer{
		cfg:        cfg,
		chunkStore: chunkStore,
		mapper:     mapper,
		quit:       make(chan struct{}),
	}
	return &writer, nil
}

// Run initializes the writer workers
func (w *Writer) Run(ctx context.Context, inChan chan chunk.Chunk) error {
	backoff := util.NewBackoff(ctx, util.BackoffConfig{
		MaxBackoff: time.Minute * 1,
		MinBackoff: time.Second * 1,
	})
	for {
		select {
		case <-ctx.Done():
			logrus.Info("shutting down writer because context was cancelled")
			return nil
		case c, open := <-inChan:
			if !open {
				return nil
			}

			ReceivedChunks.WithLabelValues().Add(1)

			remapped, err := w.mapper.MapChunk(c)
			if err != nil {
				logrus.WithError(err).Errorln("failed to remap chunk", "err", err)

				return err
			}

			// Ensure the chunk has been encoded before persisting in order to avoid
			// bad external keys in the index entry
			if remapped.Encode() != nil {
				return err
			}

			for backoff.Ongoing() {
				err = w.chunkStore.PutOne(ctx, remapped.From, remapped.Through, remapped)
				if err != nil {
					logrus.WithError(err).WithField("retries", backoff.NumRetries()).Errorf("failed to store chunk")
					backoff.Wait()
				} else {
					backoff.Reset()
					break
				}
			}
		}
	}
}
