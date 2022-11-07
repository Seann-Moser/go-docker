package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RepoImages struct {
	Count    int          `json:"count"`
	Next     interface{}  `json:"next"`
	Previous interface{}  `json:"previous"`
	Results  []*RepoImage `json:"results"`
}

func (ri *RepoImages) GetLatest(r *Repo) string {
	for _, i := range ri.Results {
		if len(i.Images) == 0 {
			continue
		}
		return i.GenerateImageName(r)
	}
	return ""
}

type RepoImage struct {
	Creator             int              `json:"creator"`
	Id                  int              `json:"id"`
	Images              []*PlatformImage `json:"images"`
	LastUpdated         time.Time        `json:"last_updated"`
	LastUpdater         int              `json:"last_updater"`
	LastUpdaterUsername string           `json:"last_updater_username"`
	Name                string           `json:"name"`
	Repository          int              `json:"repository"`
	FullSize            int              `json:"full_size"`
	V2                  bool             `json:"v2"`
	TagStatus           string           `json:"tag_status"`
	TagLastPulled       *time.Time       `json:"tag_last_pulled"`
	TagLastPushed       time.Time        `json:"tag_last_pushed"`
	MediaType           string           `json:"media_type"`
	Digest              string           `json:"digest"`
}

func (r *RepoImage) GenerateImageName(repo *Repo) string {
	return fmt.Sprintf("%s/%s:%s", repo.Namespace, repo.Name, r.Name)
}

type PlatformImage struct {
	Architecture string      `json:"architecture"`
	Features     string      `json:"features"`
	Variant      interface{} `json:"variant"`
	Digest       string      `json:"digest"`
	Os           string      `json:"os"`
	OsFeatures   string      `json:"os_features"`
	OsVersion    interface{} `json:"os_version"`
	Size         int         `json:"size"`
	Status       string      `json:"status"`
	LastPulled   *time.Time  `json:"last_pulled"`
	LastPushed   time.Time   `json:"last_pushed"`
}

func (c *Client) GetImages(ctx context.Context, repo *Repo) (*RepoImages, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=25&page=1&ordering=last_updated", repo.Namespace, repo.Name)
	resp, err := c.request(ctx, url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	repos := &RepoImages{}
	err = json.Unmarshal(resp, repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}
