package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/requests"
)

const (
	defaultScmId = "krateo"
)

type ClientOpts struct {
	ApiBaseUrl string
	Token      string
	HttpClient *http.Client
}

// Client is a tiny Github client
type Client struct {
	apiBaseUrl string
	httpClient *http.Client
	repos      *RepoService
}

// NewClient returns a new Github Client
func NewClient(opts ClientOpts) *Client {
	res := &Client{
		apiBaseUrl: opts.ApiBaseUrl,
		httpClient: opts.HttpClient,
	}

	res.repos = newRepoService(res.httpClient, res.apiBaseUrl, opts.Token)

	return res
}

func (c *Client) Repos() *RepoService {
	return c.repos
}

type Repository struct {
	Name        string `json:"name"`
	ScmId       string `json:"scmId,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	Public      bool   `json:"public"`
}

// RepoService provides methods for creating repositories.
type RepoService struct {
	client     *http.Client
	apiBaseUrl string
	token      string
}

// newRepoService returns a new RepoService.
func newRepoService(httpClient *http.Client, apiBaseUrl, token string) *RepoService {
	return &RepoService{
		client:     httpClient,
		apiBaseUrl: apiBaseUrl,
		token:      token,
	}
}

func (s *RepoService) Create(projectKey string, opts *Repository) error {
	err := requests.URL(s.apiBaseUrl).
		Method(http.MethodPost).
		Pathf("/rest/api/1.0/%s/repos", projectKey).
		Client(s.client).
		Bearer(s.token).
		BodyJSON(map[string]any{
			"name":          opts.Name,
			"public":        opts.Public,
			"auto_init":     true,
			"defaultBranch": "main",
		}).
		AddValidator(ErrorHandler(201)).
		Fetch(context.Background())
	if err != nil {
		var e StatusError
		if errors.As(err, &e) {
			return fmt.Errorf(e.Error())
		}
		return err
	}

	return nil
}

func (s *RepoService) Exists(projectKey, slug string) (bool, error) {
	err := requests.URL(s.apiBaseUrl).
		Method(http.MethodGet).
		Pathf("/rest/api/1.0/%s/repos/%s", projectKey, slug).
		Client(s.client).
		Bearer(s.token).
		AddValidator(ErrorHandler(200)).
		Fetch(context.Background())
	if err != nil {
		var e StatusError
		if errors.As(err, &e) {
			return false, fmt.Errorf(e.Error())
		}
		return false, err
	}

	return true, nil
}
