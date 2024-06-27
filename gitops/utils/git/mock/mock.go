package mock

import (
	_ "go.uber.org/mock/mockgen/model"
)

//go:generate mockgen -destination=mock_kube.go -package=mock github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git GitClient
