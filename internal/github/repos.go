package github

import (
	"fmt"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
)

type repoOwner struct {
	Login string `json:"login"`
}

type repo struct {
	Name     string    `json:"name"`
	Fork     bool      `json:"fork"`
	Archived bool      `json:"archived"`
	Private  bool      `json:"private"`
	Owner    repoOwner `json:"owner"`
}

// RepoInfo holds information about a repository.
type RepoInfo struct {
	Name    string
	Fork    bool
	Private bool
}

// newClient creates a REST client for the given host.
// If host is empty (or "github.com"), the default gh host is used.
func newClient(host string) (*api.RESTClient, error) {
	if host != "" && host != "github.com" {
		return api.NewRESTClient(api.ClientOptions{Host: host})
	}
	return api.DefaultRESTClient()
}

// ListRepos fetches all repositories for the given owner (org or user).
// It first tries the org endpoint, then falls back to the user endpoint.
// If host is empty, the default gh host is used.
func ListRepos(owner, host string, includeArchived bool) ([]RepoInfo, error) {
	client, err := newClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	return listReposWithClient(client, owner, includeArchived)
}

// CurrentUser returns the login of the authenticated user.
func CurrentUser(host string) (string, error) {
	client, err := newClient(host)
	if err != nil {
		return "", fmt.Errorf("failed to create API client: %w", err)
	}
	return currentUserWithClient(client)
}

func currentUserWithClient(client *api.RESTClient) (string, error) {
	var u struct {
		Login string `json:"login"`
	}
	if err := client.Get("user", &u); err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	return u.Login, nil
}

// ListUserOrgs fetches the organizations the authenticated user belongs to.
func ListUserOrgs(host string) ([]string, error) {
	client, err := newClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	return listUserOrgsWithClient(client)
}

func listUserOrgsWithClient(client *api.RESTClient) ([]string, error) {
	var orgs []string
	page := 1
	for {
		var pageOrgs []struct {
			Login string `json:"login"`
		}
		path := fmt.Sprintf("user/orgs?per_page=100&page=%d", page)
		if err := client.Get(path, &pageOrgs); err != nil {
			return nil, fmt.Errorf("failed to list organizations: %w", err)
		}

		if len(pageOrgs) == 0 {
			break
		}
		for _, o := range pageOrgs {
			orgs = append(orgs, o.Login)
		}

		if len(pageOrgs) < 100 {
			break
		}
		page++
	}
	return orgs, nil
}

func listReposWithClient(client *api.RESTClient, owner string, includeArchived bool) ([]RepoInfo, error) {
	repos, err := fetchRepos(client, fmt.Sprintf("orgs/%s/repos", owner))
	if err != nil {
		// Try authenticated user endpoint (includes private repos),
		// then fall back to public user endpoint
		repos, err = fetchAuthUserRepos(client, owner)
		if err != nil {
			repos, err = fetchRepos(client, fmt.Sprintf("users/%s/repos", owner))
			if err != nil {
				return nil, fmt.Errorf("failed to list repos for %s: %w", owner, err)
			}
		}
	}

	var result []RepoInfo
	for _, r := range repos {
		if !includeArchived && r.Archived {
			continue
		}
		result = append(result, RepoInfo{Name: r.Name, Fork: r.Fork, Private: r.Private})
	}
	return result, nil
}

// fetchAuthUserRepos fetches repos for the authenticated user filtered by owner.
// GET /user/repos includes private repos, unlike GET /users/{username}/repos.
func fetchAuthUserRepos(client *api.RESTClient, owner string) ([]repo, error) {
	var allRepos []repo
	page := 1

	for {
		var pageRepos []repo
		path := fmt.Sprintf("user/repos?affiliation=owner&per_page=100&page=%d", page)

		err := client.Get(path, &pageRepos)
		if err != nil {
			return nil, err
		}

		if len(pageRepos) == 0 {
			break
		}

		allRepos = append(allRepos, pageRepos...)

		if len(pageRepos) < 100 {
			break
		}
		page++
	}

	// Filter to repos owned by the specified owner
	var filtered []repo
	for _, r := range allRepos {
		if r.Owner.Login == owner {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no repos found for authenticated user %s", owner)
	}
	return filtered, nil
}

func fetchRepos(client *api.RESTClient, endpoint string) ([]repo, error) {
	var allRepos []repo
	page := 1

	for {
		var pageRepos []repo
		path := fmt.Sprintf("%s?per_page=100&page=%d", endpoint, page)

		err := client.Get(path, &pageRepos)
		if err != nil {
			var httpErr api.HTTPError
			if isHTTPError(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				return nil, err
			}
			return nil, err
		}

		if len(pageRepos) == 0 {
			break
		}

		allRepos = append(allRepos, pageRepos...)

		if len(pageRepos) < 100 {
			break
		}
		page++
	}

	return allRepos, nil
}

func isHTTPError(err error, target *api.HTTPError) bool {
	if httpErr, ok := err.(*api.HTTPError); ok {
		*target = *httpErr
		return true
	}
	return false
}
