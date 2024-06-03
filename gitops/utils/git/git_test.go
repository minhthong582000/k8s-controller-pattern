package git

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testCases = []struct {
		name        string
		gitClient   *gitClient
		url         string
		expectedOut string
		expectedErr string
	}{
		{
			name: "Clone repository",
			gitClient: &gitClient{
				token: "",
			},
			url:         "https://github.com/minhthong582000/k8s-controller-pattern.git",
			expectedOut: "",
			expectedErr: "",
		},
		{
			name: "Clone unexisted repository",
			gitClient: &gitClient{
				token: "",
			},
			url:         "https://github.com/minhthong582000/unexisted-repository.git",
			expectedOut: "",
			expectedErr: "failed to clone repository: authentication required",
		},
	}
)

func TestGitClient_CloneOrFetch_CleanUp(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			path := path.Join(os.TempDir(), strings.Replace(tt.url, "/", "_", -1))

			g := NewGitClient(tt.gitClient.token)
			err := g.CloneOrFetch(tt.url, path)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
			}

			// Call CloneOrFetch again to see if it fetches the latest changes
			err = g.CloneOrFetch(tt.url, path)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}
			assert.DirExists(t, path)

			// Clean up
			err = g.CleanUp(path)
			assert.NoError(t, err)
			assert.NoDirExists(t, path)
		})
	}
}
