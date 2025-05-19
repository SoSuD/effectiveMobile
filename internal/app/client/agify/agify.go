package agify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Agify struct {
	client  *http.Client
	baseUrl string
}

func New() *Agify {
	c := &Agify{}
	c.baseUrl = "https://api.agify.io"
	c.client = &http.Client{}
	return c
}

func (c *Agify) Get(name string) (Response, error) {
	requrl := c.baseUrl + "/?name=" + name
	req, err := http.NewRequest(http.MethodGet, requrl, nil)
	if err != nil {
		return Response{}, fmt.Errorf("agify Get Error request: %w", err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("agify Get Error request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return Response{}, fmt.Errorf(
			"unexpected status %d: %s",
			res.StatusCode, string(body),
		)
	}
	var agifyResp Response
	if err := json.NewDecoder(res.Body).Decode(&agifyResp); err != nil {
		return Response{}, fmt.Errorf("decoding JSON: %w", err)
	}
	return agifyResp, nil
}
