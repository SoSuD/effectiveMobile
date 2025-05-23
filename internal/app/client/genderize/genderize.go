package genderize

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

type Genderize struct {
	client *resty.Client
}

// New создаёт новый Genderize-клиент
func New(url string) *Genderize {
	return &Genderize{
		client: resty.New().
			SetBaseURL(url).
			SetHeader("Accept", "application/json"),
	}
}

func (c *Genderize) Get(name string) (Response, error) {
	var result Response

	resp, err := c.client.R().
		SetQueryParam("name", name).
		SetResult(&result).
		Get("/")

	if err != nil {
		return Response{}, fmt.Errorf("genderize Get error: %w", err)
	}
	if resp.IsError() {
		return Response{}, fmt.Errorf(
			"genderize Get unexpected status %d: %s",
			resp.StatusCode(), resp.String(),
		)
	}

	return result, nil
}
