package bitbucket

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/carlmjohnson/requests"
)

func TestGetRepos(t *testing.T) {
	co := &ClientOpts{
		ApiBaseUrl: "http://10.99.99.37:7990",
		Token:      "Mjg1NjM3MzA2NDIyOrGyf3Jw9/pUcoOFwy0uzQtsECox",
	}

	repos := NewClient(co).Repos()

	projectKey := "JXP"
	res, err := repos.Create(CreateRepoOpts{
		Name:       "Test Krateo 1",
		Public:     false,
		ProjectKey: projectKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("message", "first commit")
	bodyWriter.WriteField("branch", "main")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="content"; filename="README.md"`)
	h.Set("Content-Type", "application/octet-stream")
	part, err := bodyWriter.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(`# Hello`)); err != nil {
		t.Fatal(err)
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	var s string
	err = requests.URL(co.ApiBaseUrl).
		Method(http.MethodPut).
		Pathf("/rest/api/1.0/projects/%s/repos/%s/browse/README.md", projectKey, res.Slug).
		Bearer(co.Token).
		ContentType(contentType).
		BodyBytes(bodyBuf.Bytes()).
		ToString(&s).
		//AddValidator(ErrorHandler(200)).
		Fetch(context.Background())
	if err != nil {
		var e StatusError
		if errors.As(err, &e) {
			t.Fatal(e.Error())
		}
		t.Fatal(err)
	}

	t.Log(res)
}

func TestRepoAlreadyExists(t *testing.T) {
	dat, err := ioutil.ReadFile("../../../testdata/repo-already-exists.json")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusConflict)
		rw.Write(dat)
	}))
	defer server.Close()

	co := &ClientOpts{
		ApiBaseUrl: server.URL,
		HttpClient: server.Client(),
	}

	projectKey := ""
	_, err = NewClient(co).Repos().Create(CreateRepoOpts{ProjectKey: projectKey, Name: "test-krateo-1"})
	if err == nil {
		t.Fatalf("expecting an error, got nil")
	}

	if want, got := "Repository already exist.", err.Error(); got != want {
		t.Fatalf("expecting error[%s], got [%s]", want, got)
	}
}

func TestRepoCreateSuccess(t *testing.T) {
	dat, err := ioutil.ReadFile("../../../testdata/repo-create-ok.json")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusCreated)
		rw.Write(dat)
	}))
	defer server.Close()

	co := &ClientOpts{
		ApiBaseUrl: server.URL,
		HttpClient: server.Client(),
	}

	projectKey := "xxx"
	_, err = NewClient(co).Repos().Create(CreateRepoOpts{ProjectKey: projectKey, Name: "test-krateo-1"})
	if err != nil {
		t.Fatal(err)
	}
}
