package warera

import (
	"context"
	"log/slog"
	"time"
)

func (c *Client) addRequestToQueue(
	request *QueuedRequest,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queue = append(c.queue, request)
}

// takeBatchFromQueue retrieves a batch of requests from the queue.
//
// Callers should call Done() on the returned batch to ensure that all requests are properly marked as completed.
func (c *Client) takeBatchFromQueue() queueBatch {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.queue) == 0 {
		return queueBatch{}
	}

	res := queueBatch{
		Valid:   make([]*QueuedRequest, 0, c.batchSize),
		Invalid: make([]*QueuedRequest, 0, c.batchSize),
	}

	var endIdx int

	for endIdx = 0; endIdx < len(c.queue); endIdx++ {
		if c.queue[endIdx] == nil {
			continue
		}

		item := c.queue[endIdx]
		c.queue[endIdx] = nil

		if item.Ctx.Err() == nil {
			res.Valid = append(res.Valid, item)
		} else {
			res.Invalid = append(res.Invalid, item)
		}

		if len(res.Valid) >= c.batchSize {
			break
		}
	}

	if endIdx+1 < len(c.queue) {
		c.queue = c.queue[endIdx+1:]
	} else {
		c.queue = nil
	}

	return res
}

func (c *Client) wakeUpQueue() {
	select {
	case c.wakeUpQueueCh <- struct{}{}:
	default:
	}
}

func (c *Client) serveQueue(ctx context.Context) error {
	done := ctx.Done()

	for {
		select {
		case <-done:
			return nil
		case <-c.wakeUpQueueCh:
			for {
				hasNext := c.execSingleBatch(ctx)
				if !hasNext {
					break
				}
			}
		}
	}
}

func (c *Client) execSingleBatch(
	ctx context.Context,
) bool {
	batch := c.takeBatchFromQueue()
	defer batch.Done()

	if len(batch.Valid) == 0 {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err := c.execBatchRequest(ctx, batch.Valid)
	if err != nil {
		c.logger.Error("Failed to execute batch request",
			slog.Any("error", err),
		)

		// propagate the error to all requests in the batch.
		for _, req := range batch.Valid {
			if req.ErrorRef != nil {
				*req.ErrorRef = err
			}
		}
	}

	return true
}
