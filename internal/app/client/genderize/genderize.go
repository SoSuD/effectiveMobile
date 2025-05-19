package genderize

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Genderize struct {
	client  *http.Client
	baseUrl string
}

func New() *Genderize {
	c := &Genderize{}
	c.baseUrl = "https://api.genderize.io"
	c.client = &http.Client{}
	return c
}

func (c *Genderize) Get(name string) (Response, error) {
	requrl := c.baseUrl + "/?name=" + name
	req, err := http.NewRequest(http.MethodGet, requrl, nil)
	if err != nil {
		return Response{}, fmt.Errorf("genderize Get Error request: %w", err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("genderize Get Error request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return Response{}, fmt.Errorf(
			"unexpected status %d: %s",
			res.StatusCode, string(body),
		)
	}
	var nationalizeResp Response
	if err := json.NewDecoder(res.Body).Decode(&nationalizeResp); err != nil {
		return Response{}, fmt.Errorf("decoding JSON: %w", err)
	}
	return nationalizeResp, nil
}
