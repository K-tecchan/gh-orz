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
	result, err := listReposWithClient(client, "my-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d repos, want 2", len(result))
	}
	if result[0].Name != "repo-a" || result[1].Name != "repo-b" {
		t.Errorf("got %v, want [repo-a repo-b]", result)
	}
}

func TestListReposWithClient_FallbackToAuthUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/my-user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	})
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{
			{Name: "public-repo", Private: false, Owner: repoOwner{Login: "my-user"}},
			{Name: "secret-repo", Private: true, Owner: repoOwner{Login: "my-user"}},
			{Name: "other-repo", Private: false, Owner: repoOwner{Login: "other-user"}},
		}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listReposWithClient(client, "my-user", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d repos, want 2", len(result))
	}
	if result[0].Name != "public-repo" || result[0].Private != false {
		t.Errorf("result[0] = %+v, want {Name:public-repo Private:false}", result[0])
	}
	if result[1].Name != "secret-repo" || result[1].Private != true {
		t.Errorf("result[1] = %+v, want {Name:secret-repo Private:true}", result[1])
	}
}

func TestListReposWithClient_FallbackToUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/some-user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	})
	// /user/repos returns no matching repos for this owner
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]repo{})
	})
	mux.HandleFunc("/users/some-user/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{{Name: "personal-repo"}}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listReposWithClient(client, "some-user", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 1 || result[0].Name != "personal-repo" {
		t.Errorf("got %v, want [{personal-repo}]", result)
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
	result, err := listReposWithClient(client, "empty-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("got %d repos, want 0", len(result))
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
	result, err := listReposWithClient(client, "my-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d repos, want 2", len(result))
	}
	if result[0].Name != "active-repo" || result[1].Name != "another-active" {
		t.Errorf("got %v, want [active-repo another-active]", result)
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
	result, err := listReposWithClient(client, "my-org", true)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d repos, want 2", len(result))
	}
	if result[0].Name != "active-repo" || result[1].Name != "old-repo" {
		t.Errorf("got %v, want [active-repo old-repo]", result)
	}
}

func TestListReposWithClient_PrivateInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/my-org/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{
			{Name: "public-repo", Private: false},
			{Name: "private-repo", Private: true},
		}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listReposWithClient(client, "my-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d repos, want 2", len(result))
	}
	if result[0].Private != false {
		t.Errorf("result[0].Private = true, want false")
	}
	if result[1].Private != true {
		t.Errorf("result[1].Private = false, want true")
	}
}

func TestListReposWithClient_ForkInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/my-org/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []repo{
			{Name: "original-repo", Fork: false},
			{Name: "forked-repo", Fork: true},
		}
		json.NewEncoder(w).Encode(repos)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listReposWithClient(client, "my-org", false)
	if err != nil {
		t.Fatalf("listReposWithClient() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d repos, want 2", len(result))
	}
	if result[0].Fork != false {
		t.Errorf("result[0].Fork = true, want false")
	}
	if result[1].Fork != true {
		t.Errorf("result[1].Fork = false, want true")
	}
}
