package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Repositories struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []*Repo     `json:"results"`
}
type Repo struct {
	Name           string    `json:"name"`
	Namespace      string    `json:"namespace"`
	RepositoryType string    `json:"repository_type"`
	Status         int       `json:"status"`
	IsPrivate      bool      `json:"is_private"`
	StarCount      int       `json:"star_count"`
	PullCount      int       `json:"pull_count"`
	LastUpdated    time.Time `json:"last_updated"`
	DateRegistered time.Time `json:"date_registered"`
	Affiliation    string    `json:"affiliation"`
	MediaTypes     []string  `json:"media_types"`
}

func (c *Client) GetRepositories(ctx context.Context) (*Repositories, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/?page_size=25&page=1&ordering=last_updated", c.Login.Username)
	resp, err := c.request(ctx, url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	repos := &Repositories{}
	err = json.Unmarshal(resp, repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}
