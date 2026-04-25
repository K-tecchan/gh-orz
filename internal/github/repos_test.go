package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
)

// redirectTransport redirects all requests to the test server.
type redirectTransport struct {
	server *httptest.Server
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = rt.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(req)
}

func newTestClient(t *testing.T, server *httptest.Server) *api.RESTClient {
	t.Helper()
	client, err := api.NewRESTClient(api.ClientOptions{
		Host:      "github.com",
		AuthToken: "test-token",
		Transport: &redirectTransport{server: server},
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestListReposWithClient_OrgEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/my-org/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{{Name: "repo-a"}, {Name: "repo-b"}}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	names, err := listReposWithClient(client, "my-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("got %d repos, want 2", len(names))
	}
	if names[0] != "repo-a" || names[1] != "repo-b" {
		t.Errorf("got %v, want [repo-a repo-b]", names)
	}
}

func TestListReposWithClient_FallbackToUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/some-user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	})
	mux.HandleFunc("/users/some-user/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{{Name: "personal-repo"}}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	names, err := listReposWithClient(client, "some-user", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(names) != 1 || names[0] != "personal-repo" {
		t.Errorf("got %v, want [personal-repo]", names)
	}
}

func TestListReposWithClient_EmptyRepos(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/empty-org/repos", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]repo{})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	names, err := listReposWithClient(client, "empty-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("got %d repos, want 0", len(names))
	}
}

func TestListReposWithClient_ExcludesArchived(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/my-org/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{
			{Name: "active-repo", Archived: false},
			{Name: "old-repo", Archived: true},
			{Name: "another-active", Archived: false},
		}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	names, err := listReposWithClient(client, "my-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("got %d repos, want 2", len(names))
	}
	if names[0] != "active-repo" || names[1] != "another-active" {
		t.Errorf("got %v, want [active-repo another-active]", names)
	}
}

func TestListReposWithClient_IncludesArchived(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/my-org/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{
			{Name: "active-repo", Archived: false},
			{Name: "old-repo", Archived: true},
		}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	names, err := listReposWithClient(client, "my-org", true)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("got %d repos, want 2", len(names))
	}
	if names[0] != "active-repo" || names[1] != "old-repo" {
		t.Errorf("got %v, want [active-repo old-repo]", names)
	}
}
