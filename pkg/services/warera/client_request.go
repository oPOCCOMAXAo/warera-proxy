package warera

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/opoccomaxao/warera-proxy/pkg/apperr"
	pkgerr "github.com/pkg/errors"
)

// RawRequest dogoc
//
// Returns raw json response or errors:
//   - [apperr.ExportedError] - for public error return;
//   - other errors - for logging.
func (c *Client) RawRequest(
	ctx context.Context,
	method string,
	params json.RawMessage,
) (json.RawMessage, error) {
	var (
		err error
		res json.RawMessage
	)

	qreq := &QueuedRequest{
		Ctx: ctx,
		WG:  sync.WaitGroup{},

		Method:    method,
		Params:    params,
		ResultRef: &res,
		ErrorRef:  &err,
	}

	qreq.WG.Add(1)
	c.addRequestToQueue(qreq)
	c.wakeUpQueue()
	qreq.WG.Wait()

	return res, err
}

//nolint:cyclop,funlen
func (c *Client) execBatchRequest(
	ctx context.Context,
	requests []*QueuedRequest,
) error {
	if len(requests) == 0 {
		return nil
	}

	err := c.rateLimit.Wait(ctx)
	if err != nil {
		return pkgerr.WithStack(apperr.NewExportedError(err))
	}

	methods := make([]string, len(requests))
	params := make(map[string]json.RawMessage, len(requests))
	results := []json.RawMessage{}

	for i, req := range requests {
		methods[i] = req.Method
		if req.Params != nil {
			params[strconv.Itoa(i)] = req.Params
		}
	}

	urlFull := c.host + "/trpc/" + strings.Join(methods, ",") + "?batch=1"

	var body io.Reader = http.NoBody

	if len(params) > 0 {
		jsonBytes, err := json.Marshal(params)
		if err != nil {
			return pkgerr.WithStack(apperr.NewExportedError(err))
		}

		body = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlFull, body)
	if err != nil {
		return pkgerr.WithStack(apperr.NewExportedError(err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.token)

	res, err := c.client.Do(req)
	if err != nil {
		return pkgerr.WithStack(apperr.NewExportedError(err))
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return pkgerr.WithStack(apperr.NewExportedError(err,
			slog.Int("status", res.StatusCode),
		))
	}

	rawRes, err := io.ReadAll(res.Body)
	if err != nil {
		return pkgerr.WithStack(apperr.NewExportedError(err))
	}

	err = json.Unmarshal(rawRes, &results)
	if err != nil {
		return pkgerr.WithStack(apperr.NewExportedError(err))
	}

	// ensure results has enough capacity to avoid out of range errors in case the response is shorter than the requests.
	results = slices.Grow(results, len(requests))

	for i, request := range requests {
		if request.ResultRef == nil {
			continue
		}

		if request.ResultRef != nil {
			*request.ResultRef = results[i]
		}
	}

	return nil
}
