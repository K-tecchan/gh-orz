package github

import (
	"encoding/json"
	"fmt"
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

func TestCurrentUserWithClient(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			Login string `json:"login"`
		}{Login: "my-user"})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	login, err := currentUserWithClient(client)
	if err != nil {
		t.Fatalf("currentUserWithClient() error: %v", err)
	}
	if login != "my-user" {
		t.Errorf("got %q, want %q", login, "my-user")
	}
}

func TestCurrentUserWithClient_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Bad credentials"}`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	if _, err := currentUserWithClient(client); err == nil {
		t.Fatal("currentUserWithClient() error = nil, want error")
	}
}

func TestListUserOrgsWithClient(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		orgs := []struct {
			Login string `json:"login"`
		}{{Login: "org-a"}, {Login: "org-b"}}
		json.NewEncoder(w).Encode(orgs)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listUserOrgsWithClient(client)
	if err != nil {
		t.Fatalf("listUserOrgsWithClient() error: %v", err)
	}

	if len(result) != 2 || result[0] != "org-a" || result[1] != "org-b" {
		t.Errorf("got %v, want [org-a org-b]", result)
	}
}

func TestListUserOrgsWithClient_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		var orgs []struct {
			Login string `json:"login"`
		}
		switch r.URL.Query().Get("page") {
		case "1":
			for i := range 100 {
				orgs = append(orgs, struct {
					Login string `json:"login"`
				}{Login: fmt.Sprintf("org-%d", i)})
			}
		case "2":
			orgs = append(orgs, struct {
				Login string `json:"login"`
			}{Login: "org-100"})
		}
		json.NewEncoder(w).Encode(orgs)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listUserOrgsWithClient(client)
	if err != nil {
		t.Fatalf("listUserOrgsWithClient() error: %v", err)
	}

	if len(result) != 101 {
		t.Fatalf("got %d orgs, want 101", len(result))
	}
	if result[0] != "org-0" || result[99] != "org-99" || result[100] != "org-100" {
		t.Errorf("unexpected org order: first=%q last-of-page1=%q page2=%q", result[0], result[99], result[100])
	}
}

func TestListUserOrgsWithClient_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]struct {
			Login string `json:"login"`
		}{})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	result, err := listUserOrgsWithClient(client)
	if err != nil {
		t.Fatalf("listUserOrgsWithClient() error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d orgs, want 0", len(result))
	}
}

func TestListUserOrgsWithClient_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal Server Error"}`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	if _, err := listUserOrgsWithClient(client); err == nil {
		t.Fatal("listUserOrgsWithClient() error = nil, want error")
	}
}
