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
		name      string
		gitClient *gitClient
		url       string
		want      string
		err       string
	}{
		{
			name: "Clone repository",
			gitClient: &gitClient{
				token: "",
			},
			url:  "https://github.com/minhthong582000/k8s-controller-pattern.git",
			want: "",
			err:  "",
		},
		{
			name: "Clone unexisted repository",
			gitClient: &gitClient{
				token: "",
			},
			url:  "https://github.com/minhthong582000/unexisted-repository.git",
			want: "",
			err:  "failed to clone repository: authentication required",
		},
	}
)

func TestGitClient_CloneOrFetch_CleanUp(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			path := path.Join(os.TempDir(), strings.Replace(tt.url, "/", "_", -1))

			g := &gitClient{
				token: tt.gitClient.token,
			}
			err := g.CloneOrFetch(tt.url, path)
			if err != nil {
				if tt.err != err.Error() {
					assert.Equal(t, tt.err, err)
				}
			}

			// Call CloneOrFetch again to see if it fetches the latest changes
			err = g.CloneOrFetch(tt.url, path)
			if err != nil {
				if tt.err != err.Error() {
					assert.Equal(t, tt.err, err)
				}
			}

			// Clean up
			err = g.CleanUp(path)
			assert.Nil(t, err)
		})
	}
}
