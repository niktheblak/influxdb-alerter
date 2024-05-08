package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	gohttp "net/http"
	"time"
)

type Config struct {
	URL           string
	Method        string
	Authorization string
	ContentType   string
	Encode        func(message string) []byte
}

type HTTP struct {
	cfg Config
}

func New(cfg Config) *HTTP {
	return &HTTP{cfg: cfg}
}

func (h *HTTP) Notify(ctx context.Context, message string) error {
	client := &gohttp.Client{}
	dl, ok := ctx.Deadline()
	if ok {
		client.Timeout = dl.Sub(time.Now())
	} else {
		client.Timeout = 10 * time.Second
	}
	var msg []byte
	if h.cfg.Encode != nil {
		msg = h.cfg.Encode(message)
	} else {
		msg = []byte(message)
	}
	req, err := gohttp.NewRequest(h.cfg.Method, h.cfg.URL, bytes.NewReader(msg))
	if err != nil {
		return err
	}
	if h.cfg.Authorization != "" {
		req.Header.Set("Authorization", h.cfg.Authorization)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("http server responded with error %d: %s", resp.StatusCode, string(body)))
	}
	return nil
}
