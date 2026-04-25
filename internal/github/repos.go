package github

import (
	"fmt"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
)

type repo struct {
	Name     string `json:"name"`
	Fork     bool   `json:"fork"`
	Archived bool   `json:"archived"`
}

// RepoInfo holds information about a repository.
type RepoInfo struct {
	Name string
	Fork bool
}

// ListRepos fetches all repositories for the given owner (org or user).
// It first tries the org endpoint, then falls back to the user endpoint.
// If host is empty, the default gh host is used.
func ListRepos(owner, host string, includeArchived bool) ([]RepoInfo, error) {
	var client *api.RESTClient
	var err error
	if host != "" && host != "github.com" {
		client, err = api.NewRESTClient(api.ClientOptions{Host: host})
	} else {
		client, err = api.DefaultRESTClient()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	return listReposWithClient(client, owner, includeArchived)
}

func listReposWithClient(client *api.RESTClient, owner string, includeArchived bool) ([]RepoInfo, error) {
	repos, err := fetchRepos(client, fmt.Sprintf("orgs/%s/repos", owner))
	if err != nil {
		// Fall back to user endpoint
		repos, err = fetchRepos(client, fmt.Sprintf("users/%s/repos", owner))
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for %s: %w", owner, err)
		}
	}

	var result []RepoInfo
	for _, r := range repos {
		if !includeArchived && r.Archived {
			continue
		}
		result = append(result, RepoInfo{Name: r.Name, Fork: r.Fork})
	}
	return result, nil
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
