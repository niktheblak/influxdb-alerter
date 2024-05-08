package notifier

import (
	"context"
)

type Notifier interface {
	Notify(ctx context.Context, message string) error
}
