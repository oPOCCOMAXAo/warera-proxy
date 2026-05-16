package warera

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"golang.org/x/time/rate"
)

var _ lifecycle.Servable = (*Client)(nil)

type Client struct {
	mu sync.Mutex

	token string
	host  string

	logger    *slog.Logger
	client    *http.Client
	rateLimit *rate.Limiter
	batchSize int

	queue         []*QueuedRequest
	wakeUpQueueCh chan struct{}
}

//nolint:mnd
func NewClient(
	config Config,
	logger *slog.Logger,
) *Client {
	res := &Client{
		token: config.Token,
		host:  config.Host,

		logger: logger,

		client: &http.Client{
			Timeout: 60 * time.Second,
		},

		rateLimit: rate.NewLimiter(rate.Limit(config.RequestsPerMinute)/60.0, 1),
		batchSize: config.BatchSize,

		wakeUpQueueCh: make(chan struct{}, 2),
	}

	return res
}

func (c *Client) Serve(ctx context.Context) error {
	c.logger.Info("Started")

	return c.serveQueue(ctx)
}
