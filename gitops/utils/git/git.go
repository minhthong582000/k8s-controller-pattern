package git

import (
	"fmt"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitClient interface {
	CloneOrFetch(url, path string) error
	Checkout(path, revision string) (string, error)
	CleanUp(path string) error
}

type gitClient struct {
	token string
}

func NewGitClient(token string) GitClient {
	return &gitClient{
		token: token,
	}
}

func (g *gitClient) CloneOrFetch(url, path string) error {
	// Need to clone the repository
	if _, err := os.Stat(path); os.IsNotExist(err) {
		var authOption *http.BasicAuth

		if g.token == "" {
			authOption = nil
		} else {
			// The intended use of a GitHub personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			authOption = &http.BasicAuth{
				Username: "github", // yes, this can be anything except an empty string
				Password: g.token,
			}
		}

		_, err := git.PlainClone(path, false, &git.CloneOptions{
			Auth: authOption,
			URL:  url,
		})
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		return nil
	}

	// Fetch the latest changes if it's already cloned
	r, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	err = r.Fetch(&git.FetchOptions{
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	return nil
}

func (g *gitClient) Checkout(path, revision string) (string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(plumbing.NewBranchReferenceName(revision)),
		Force:  true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to checkout revision: %w", err)
	}

	// Git pull
	err = w.Pull(&git.PullOptions{
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("failed to pull: %w", err)
	}

	// Find current HEAD
	ref, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return ref.Hash().String(), nil
}

func (g *gitClient) CleanUp(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to clean up repository: %w", err)
	}

	return nil
}
