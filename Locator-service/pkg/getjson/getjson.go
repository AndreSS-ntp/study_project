package getjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func GetJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return fmt.Errorf("error while getting URL from client: %w", errors.New("not found"))
	}
	defer func() {
		def_err := r.Body.Close()
		if def_err != nil {
			def_err = fmt.Errorf("error while closing response body: %w", def_err)
			fmt.Println(def_err)
		}
	}()
	return json.NewDecoder(r.Body).Decode(target)
}
