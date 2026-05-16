package warera

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/opoccomaxao/warera-proxy/pkg/utils/jsonutils"
	"github.com/tidwall/gjson"
)

type ProxyRequest struct {
	// Methods - comma-separated methods.
	Methods string

	// Batch query parameter, which changes request to batch processing.
	Batch string

	// Input raw input.
	Input []byte

	isJSONValid bool
}

type ProxyResponse struct {
	RawBody json.RawMessage

	HTTPStatus int
}

func (c *Client) ProxyRequest(
	ctx context.Context,
	request ProxyRequest,
) *ProxyResponse {
	request.isJSONValid = json.Valid(request.Input)

	if request.Batch == "1" {
		return c.proxyBatchRequest(ctx, request)
	}

	return c.proxySingleRequest(ctx, request)
}

func (c *Client) proxyBatchRequest(
	ctx context.Context,
	request ProxyRequest,
) *ProxyResponse {
	methods, params, errRes := c.parseBatchRequest(request)
	if errRes != nil {
		return errRes
	}

	responses := c.execProxyBatchRequest(ctx, methods, params)

	return c.makeProxyBatchResponse(responses)
}

func (c *Client) parseBatchRequest(
	request ProxyRequest,
) ([]string, map[int]json.RawMessage, *ProxyResponse) {
	methods := strings.Split(request.Methods, ",")
	params := make(map[int]json.RawMessage, len(methods))

	var errText string

	if request.isJSONValid {
		err := json.Unmarshal(request.Input, &params)
		if err == nil {
			return methods, params, nil
		}

		errText = err.Error()
	} else {
		errText = TextErrorInvalidJSON
	}

	res := &Response{
		Error: &Error{
			Message: "proxy: " + errText,
			Code:    ErrorCodeInvalidJSON,
			Data: ErrorData{
				Code:       TextErrorCodeBadRequest,
				HTTPStatus: http.StatusBadRequest,
				Path:       "",
			},
		},
	}

	results := make(map[int]json.RawMessage, len(methods))

	for i, method := range methods {
		res.Error.Data.Path = method

		results[i] = jsonutils.MarshalNoError(res)
	}

	return nil, nil, &ProxyResponse{
		RawBody:    jsonutils.MarshalNoError(results),
		HTTPStatus: http.StatusBadRequest,
	}
}

func (c *Client) execProxyBatchRequest(
	ctx context.Context,
	methods []string,
	params map[int]json.RawMessage,
) []*ProxyResponse {
	res := make([]*ProxyResponse, len(methods))

	var wg sync.WaitGroup

	for i, method := range methods {
		wg.Go(func() {
			res[i] = c.proxySingleRequest(ctx, ProxyRequest{
				Methods:     method,
				Input:       params[i],
				isJSONValid: true,
			})
		})
	}

	wg.Wait()

	return res
}

func (c *Client) makeProxyBatchResponse(
	responses []*ProxyResponse,
) *ProxyResponse {
	results := make(map[int]json.RawMessage, len(responses))
	statuses := make(map[int]struct{}, len(responses))

	for i, response := range responses {
		results[i] = response.RawBody
		statuses[response.HTTPStatus] = struct{}{}
	}

	res := &ProxyResponse{
		RawBody: jsonutils.MarshalNoError(results),
	}

	if len(statuses) == 1 {
		for status := range statuses {
			res.HTTPStatus = status

			break
		}
	} else {
		res.HTTPStatus = http.StatusMultiStatus
	}

	return res
}

func (c *Client) proxySingleRequest(
	ctx context.Context,
	request ProxyRequest,
) *ProxyResponse {
	if !request.isJSONValid {
		return &ProxyResponse{
			RawBody: jsonutils.MarshalNoError(&Response{
				Error: &Error{
					Message: "proxy: " + TextErrorInvalidJSON,
					Code:    ErrorCodeInvalidJSON,
					Data: ErrorData{
						Code:       TextErrorCodeBadRequest,
						HTTPStatus: http.StatusBadRequest,
						Path:       request.Methods,
					},
				},
			}),
			HTTPStatus: http.StatusBadRequest,
		}
	}

	rawRes, err := c.RawRequest(ctx, request.Methods, request.Input)
	if err != nil {
		return &ProxyResponse{
			RawBody:    ProxyError(err.Error()),
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	res := &ProxyResponse{
		RawBody: rawRes,
	}

	status := gjson.GetBytes(rawRes, "error.data.httpStatus")
	if status.Exists() {
		res.HTTPStatus = int(status.Int())
	} else {
		res.HTTPStatus = http.StatusOK
	}

	return res
}
