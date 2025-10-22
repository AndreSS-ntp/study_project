package http_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/domain"
	"net/http"
	"time"
)

type HttpClient struct {
	myClient *http.Client
}

func NewHttpClient(timeout time.Duration) *HttpClient {
	return &HttpClient{&http.Client{Timeout: timeout}}
}

func (h *HttpClient) GetSystem(ctx context.Context, url string) (*domain.System, error) {
	r, err := h.myClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while getting URL from client: %w", errors.New("not found"))
	}
	defer func() {
		def_err := r.Body.Close()
		if def_err != nil {
			alogger.FromContext(ctx).Error(ctx, "error while closing response body: "+def_err.Error())
		}
	}()
	s := &domain.System{}
	err_decoder := json.NewDecoder(r.Body).Decode(s)
	if err_decoder != nil {
		alogger.FromContext(ctx).Error(ctx, "error while decoding json: "+err_decoder.Error())
	}

	return s, nil
}
