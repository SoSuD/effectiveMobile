package nationalize

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Nationalize struct {
	client *resty.Client
}

// New создаёт новый Nationalize-клиент
func New(url string) *Nationalize {
	return &Nationalize{
		client: resty.New().
			SetBaseURL(url).
			SetHeader("Accept", "application/json"),
	}
}

func (c *Nationalize) Get(name string) (Response, error) {
	var result Response

	resp, err := c.client.R().
		SetQueryParam("name", name).
		SetResult(&result).
		Get("/")

	if err != nil {
		return Response{}, fmt.Errorf("nationalize Get error: %w", err)
	}
	if resp.IsError() {
		return Response{}, fmt.Errorf(
			"nationalize Get unexpected status %d: %s",
			resp.StatusCode(), resp.String(),
		)
	}

	return result, nil
}
