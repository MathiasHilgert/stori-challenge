package mailing

import (
	"context"
	"stori-challenge/internal/summaries"
)

type Mailer interface {
	Send(ctx context.Context, to string, summary summaries.Summary) error
}
