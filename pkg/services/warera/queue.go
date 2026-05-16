package warera

import (
	"context"
	"encoding/json"
	"sync"
)

// QueuedRequest is an internal struct used to represent individual requests within a batch operation.
//
//nolint:containedctx
type QueuedRequest struct {
	Ctx context.Context
	WG  sync.WaitGroup

	Method    string
	Params    json.RawMessage
	ResultRef *json.RawMessage
	ErrorRef  *error
}

type queueBatch struct {
	Valid   []*QueuedRequest
	Invalid []*QueuedRequest
}

func (b *queueBatch) Done() {
	for _, req := range b.Valid {
		req.WG.Done()
	}

	for _, req := range b.Invalid {
		req.WG.Done()
	}
}
