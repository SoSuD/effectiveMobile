package agify

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

type Agify struct {
	client *resty.Client
}

func New(url string) *Agify {
	return &Agify{
		client: resty.New().
			SetBaseURL(url).
			SetHeader("Accept", "application/json"),
	}
}

func (c *Agify) Get(name string) (Response, error) {
	var result Response

	resp, err := c.client.R().
		SetQueryParam("name", name).
		SetResult(&result).
		Get("/")

	if err != nil {
		return Response{}, fmt.Errorf("agify Get error: %w", err)
	}

	if resp.IsError() {
		return Response{}, fmt.Errorf(
			"agify Get unexpected status %d: %s",
			resp.StatusCode(), resp.String(),
		)
	}

	return result, nil
}
