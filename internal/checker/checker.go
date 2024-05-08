package checker

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const queryTemplate = `from(bucket: "%s")
  |> range(start: -1h)
  |> filter(fn: (r) =>
      r._measurement == "%s"
  )
  |> top(n:1, columns: ["_time", "name"])`

var (
	ErrNoResults = errors.New("query returned no results")
)

type ErrNoRecentData struct {
	Since time.Time
}

func (e *ErrNoRecentData) Error() string {
	return fmt.Sprintf("no recent data since %s", e.Since)
}

// Config is the InfluxDB connection config
type Config struct {
	Addr        string
	Org         string
	Token       string
	Bucket      string
	Measurement string
	Since       time.Duration
}

type Checker interface {
	io.Closer
	Check(ctx context.Context) error
}

type checker struct {
	client   influxdb2.Client
	queryAPI api.QueryAPI
	logger   *slog.Logger
	cfg      Config
}

func New(cfg Config, logger *slog.Logger) (Checker, error) {
	client := influxdb2.NewClientWithOptions(cfg.Addr, cfg.Token, influxdb2.DefaultOptions().
		SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		}))
	return &checker{
		client:   client,
		queryAPI: client.QueryAPI(cfg.Org),
		logger:   logger,
		cfg:      cfg,
	}, nil
}

func (c *checker) Check(ctx context.Context) (err error) {
	q := fmt.Sprintf(queryTemplate, c.cfg.Bucket, c.cfg.Measurement)
	c.logger.Debug("Executing query", "query", q)
	res, err := c.queryAPI.Query(ctx, q)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := res.Close()
		err = errors.Join(err, closeErr)
	}()
	ok := res.Next()
	if !ok {
		c.logger.Warn("Query returned no results")
		err = ErrNoResults
		return
	}
	r := res.Record()
	c.logger.Debug("Query result", "record", r.String())
	_, ok = r.ValueByKey("name").(string)
	if !ok {
		err = ErrNoResults
		return
	}
	_, ok = r.ValueByKey("_field").(string)
	if !ok {
		err = ErrNoResults
		return
	}
	newest := r.Time()
	c.logger.Info("Most recent result", "ts", newest)
	now := time.Now()
	if newest.Before(now.Add(-c.cfg.Since)) {
		err = &ErrNoRecentData{Since: newest}
		return
	}
	err = res.Err()
	return
}

func (c *checker) Close() error {
	c.client.Close()
	return nil
}
