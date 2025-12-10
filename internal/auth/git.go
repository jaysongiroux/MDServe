package auth

import "github.com/go-git/go-git/v6/plumbing/transport/http"

func CreateGitBasicAuth(username *string, password *string) *http.BasicAuth {
	if username != nil && password != nil {
		return &http.BasicAuth{
			Username: *username,
			Password: *password,
		}
	}
	return nil
}
