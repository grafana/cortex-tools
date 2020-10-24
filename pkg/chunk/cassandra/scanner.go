package cassandra

import (
	"context"
	"fmt"
	"sync"

	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/cortexproject/cortex/pkg/chunk/cassandra"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

// scanBatch represents a batch of rows read from Cassandra.
type scanBatch struct {
	hash       []byte
	rangeValue []byte
	value      []byte
}

type IndexValidator struct {
	schema chunk.SchemaConfig
	s      *StorageClient
	o      *ObjectClient
}

func NewIndexValidator(
	cfg cassandra.Config,
	schema chunk.SchemaConfig,
) (*IndexValidator, error) {
	logrus.Debug("Connecting to Cassandra")
	o, err := NewObjectClient(
		cfg,
		schema,
		prometheus.NewRegistry(),
	)
	if err != nil {
		return nil, err
	}

	s, err := NewStorageClient(
		cfg,
		schema,
		prometheus.NewRegistry(),
	)
	if err != nil {
		return nil, err
	}

	logrus.Debug("Connected")
	return &IndexValidator{schema, s, o}, nil
}

func (i *IndexValidator) IndexScan(ctx context.Context, table string, from model.Time, to model.Time, out chan string) error {
	q := i.s.readSession.Query(fmt.Sprintf("SELECT hash, range, value FROM %s", table))

	iter := q.WithContext(ctx).Iter()
	defer iter.Close()
	scanner := iter.Scanner()

	wg := &sync.WaitGroup{}
	batchChan := make(chan scanBatch)

	for n := 0; n < 16; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for b := range batchChan {
				i.checkEntry(ctx, from, to, out, b)
			}
		}()
	}

	rowsReadTotal := 0

	for scanner.Next() {
		b := scanBatch{}
		if err := scanner.Scan(&b.hash, &b.rangeValue, &b.value); err != nil {
			return errors.WithStack(err)
		}
		batchChan <- b
		rowsReadTotal++
		if rowsReadTotal%1000 == 0 {
			logrus.Infof("index entries scanned total: %d\n", rowsReadTotal)
		}
	}
	close(batchChan)
	wg.Wait()
	return errors.WithStack(scanner.Err())
}

func (i *IndexValidator) checkEntry(
	ctx context.Context,
	from model.Time,
	to model.Time,
	out chan string,
	entry scanBatch,
) {
	chunkID, _, isSeriesID, err := parseChunkTimeRangeValue(entry.rangeValue, entry.value)
	if err != nil {
		logrus.WithField("chunk_id", chunkID).WithError(err).Errorln("unable to parse chunk time range value")
		return
	}

	if isSeriesID {
		logrus.WithField("series_id", chunkID).Debugln("ignoring series id row")
		return
	}

	c, err := chunk.ParseExternalKey("fake", chunkID)
	if err != nil {
		logrus.WithField("chunk_id", chunkID).WithError(err).Errorln("unable to parse external key")
		return
	}

	if from > c.Through || c.From > to {
		logrus.WithField("chunk_id", chunkID).Debugln("ignoring chunk outside time range")
		return
	}

	chunkTable, err := i.schema.ChunkTableFor(c.From)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"chunk_id": chunkID,
			"from":     c.From.String(),
			"through":  c.Through.String(),
		}).WithError(err).Errorln("unable to determine chunk table")
		return
	}

	var exists int
	err = i.o.readSession.Query(
		fmt.Sprintf("SELECT count(*) FROM %s WHERE hash = ?", chunkTable),
		c.ExternalKey(),
	).WithContext(ctx).Scan(&exists)

	if err != nil {
		fmt.Println(err)
		return
	}
	if exists == 0 {
		logrus.WithField("chunk_id", chunkID).Infoln("chunk not found, adding to output file")
		out <- fmt.Sprintf("%s,0x%x\n", string(entry.hash), entry.rangeValue)
	} else {
		logrus.WithField("chunk_id", chunkID).Debugln("chunk found")
	}
}
