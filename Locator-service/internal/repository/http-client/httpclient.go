package http_client

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (h *HttpClient) GetSystem(url string) (*domain.System, error) {
	r, err := h.myClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while getting URL from client: %w", errors.New("not found"))
	}
	defer func() {
		def_err := r.Body.Close()
		if def_err != nil {
			def_err = fmt.Errorf("error while closing response body: %w", def_err)
			fmt.Println(def_err)
		}
	}()
	s := &domain.System{}
	err_decoder := json.NewDecoder(r.Body).Decode(s)
	if err_decoder != nil {
		err_decoder = fmt.Errorf("error while decoding json: %w", err_decoder)
		fmt.Println(err_decoder)
	}

	return s, nil
}
