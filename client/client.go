package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Client struct {
	httpClient *http.Client
	Login      *Login
	Token      *Token
}

const (
	dockerUsernameFlag = "dockerhub-user-name"
	dockerApiKeyFlag   = "dockerhub-api-key"

	dockerSignInURL = "https://hub.docker.com/v2/users/login"
)

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("docker", pflag.ExitOnError)
	pflag.String(dockerUsernameFlag, "", "")
	pflag.String(dockerApiKeyFlag, "", "")
	return fs
}

func New(ctx context.Context) (*Client, error) {
	if v := viper.GetString(dockerUsernameFlag); v == "" {
		return nil, errors.New("missing data for flag:" + dockerUsernameFlag)
	}

	if v := viper.GetString(dockerApiKeyFlag); v == "" {
		return nil, errors.New("missing data for flag:" + dockerApiKeyFlag)
	}
	c := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		Login: &Login{
			Username: viper.GetString(dockerUsernameFlag),
			Password: viper.GetString(dockerApiKeyFlag),
		},
		Token: &Token{},
	}
	err := c.login(ctx)
	if err != nil {
		return nil, err
	}
	err = c.signInLocally()
	if err != nil {
		return nil, err
	}
	return c, nil
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

func (c *Client) login(ctx context.Context) error {
	body, err := json.Marshal(c.Login)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, dockerSignInURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("invalid status code %s", string(b))
	}

	return json.NewDecoder(resp.Body).Decode(c.Token)
}

func (c *Client) request(ctx context.Context, url, method string, rawBody interface{}) ([]byte, error) {
	var body []byte
	var err error
	if rawBody != nil {
		body, err = json.Marshal(rawBody)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token.Token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid status code %s", string(b))
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) Pull(ctx context.Context, repo *Repo) error {
	repoImage, err := c.GetImages(ctx, repo)
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "pull", repoImage.GetLatest(repo))
	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
func (c *Client) signInLocally() error {
	cmd := exec.Command("docker", "login", "-u", c.Login.Username, "-p", c.Login.Password)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
