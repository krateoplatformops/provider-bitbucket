package bitbucket

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/carlmjohnson/requests"
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
func NewClient(opts *ClientOpts) *Client {
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

type CreateRepoOpts struct {
	Name          string
	Public        bool
	DefaultBranch string
	ProjectKey    string
}

func (s *RepoService) Create(opts CreateRepoOpts) (*Repository, error) {
	if opts.DefaultBranch == "" {
		opts.DefaultBranch = "main"
	}

	resp := &Repository{}

	err := requests.URL(s.apiBaseUrl).
		Method(http.MethodPost).
		Pathf("/rest/api/1.0/projects/%s/repos", opts.ProjectKey).
		Client(s.client).
		Bearer(s.token).
		BodyJSON(map[string]interface{}{
			"name":          opts.Name,
			"public":        opts.Public,
			"defaultBranch": opts.DefaultBranch,
		}).
		AddValidator(ErrorHandler(201)).
		ToJSON(resp).
		Fetch(context.Background())
	if err != nil {
		var e StatusError
		if errors.As(err, &e) {
			return nil, fmt.Errorf(e.Error())
		}
		return nil, err
	}

	return resp, nil
}

type RepoInitOpts struct {
	ProjectKey string
	RepoSlug   string
	Title      string
}

func (s *RepoService) Init(opts RepoInitOpts) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("message", "first commit")
	bodyWriter.WriteField("branch", "main")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="content"; filename="README.md"`)
	h.Set("Content-Type", "application/octet-stream")
	part, err := bodyWriter.CreatePart(h)
	if err != nil {
		return err
	}

	if opts.Title == "" {
		opts.Title = opts.RepoSlug
	}
	content := []byte(fmt.Sprintf("# %s", opts.Title))
	if _, err := part.Write(content); err != nil {
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	//var res string
	err = requests.URL(s.apiBaseUrl).
		Method(http.MethodPut).
		Pathf("/rest/api/1.0/projects/%s/repos/%s/browse/README.md", opts.ProjectKey, opts.RepoSlug).
		Bearer(s.token).
		ContentType(contentType).
		BodyBytes(bodyBuf.Bytes()).
		//ToString(&res).
		Fetch(context.Background())
	if err != nil {
		var e StatusError
		if errors.As(err, &e) {
			return err
		}
		return err
	}

	return nil
}

func (s *RepoService) Exists(projectKey, slug string) (bool, error) {
	err := requests.URL(s.apiBaseUrl).
		Method(http.MethodGet).
		Pathf("/rest/api/1.0/projects/%s/repos/%s", projectKey, slug).
		Client(s.client).
		Bearer(s.token).
		AddValidator(ErrorHandler(200)).
		Fetch(context.Background())
	if err != nil {
		var e StatusError
		if errors.As(err, &e) {
			if e.Code == 404 {
				return false, nil
			}
			return false, fmt.Errorf(e.Error())
		}
		return false, err
	}

	return true, nil
}
