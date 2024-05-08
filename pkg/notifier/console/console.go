package console

import (
	"context"
	"fmt"
	"io"
	"os"
)

type Console struct {
	Writer io.Writer
}

func (c *Console) Notify(ctx context.Context, message string) error {
	if c.Writer == nil {
		c.Writer = os.Stdout
	}
	_, err := fmt.Fprintln(c.Writer, message)
	return err
}
