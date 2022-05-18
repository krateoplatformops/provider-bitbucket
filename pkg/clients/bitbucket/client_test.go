package bitbucket

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepoNotExists(t *testing.T) {
	dat, err := ioutil.ReadFile("../../../testdata/repo-not-exists.json")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write(dat)
	}))
	defer server.Close()

	co := ClientOpts{
		ApiBaseUrl: server.URL,
		HttpClient: server.Client(),
	}

	projectKey := ""
	ok, err := NewClient(co).Repos().Exists(projectKey, "test-krateo-1")
	if err == nil {
		t.Fatalf("expecting an error, got nil")
	} else if want, got := "Repository does not exist.", err.Error(); want != got {
		t.Fatalf("expecting error[%s], got [%s]", want, got)
	}

	if ok {
		t.Fatalf("excpected repository exists: false")
	}
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

	co := ClientOpts{
		ApiBaseUrl: server.URL,
		HttpClient: server.Client(),
	}

	projectKey := ""

	ro := Repository{
		Name: "test-krateo-1",
	}
	err = NewClient(co).Repos().Create(projectKey, &ro)
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

	co := ClientOpts{
		ApiBaseUrl: server.URL,
		HttpClient: server.Client(),
	}

	projectKey := "xxx"
	ro := Repository{
		Name: "test-krateo-1",
	}
	err = NewClient(co).Repos().Create(projectKey, &ro)
	if err != nil {
		t.Fatal(err)
	}
}
