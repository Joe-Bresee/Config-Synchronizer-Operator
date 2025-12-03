package source

import (
	"context"
	"fmt"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FetchSource(configSync *configsv1alpha1.ConfigSync, ctx context.Context, c client.Client) (string, string, string, error) {
	if configSync.Spec.Source.Git == nil {
		return "", "", "", fmt.Errorf("only Git sources are supported; please set spec.source.git")
	}

	revisionSHA, sourcePath, commitMsg, err := cloneOrUpdate(
		ctx,
		c,
		configSync.Spec.Source.Git.RepoURL,
		configSync.Spec.Source.Git.Revision,
		configSync.Spec.Source.Git.Branch,
		configSync.Spec.Source.Git.AuthMethod,
		configSync.Spec.Source.Git.AuthSecretRef,
	)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to clone or update git repository: %w", err)
	}

	return revisionSHA, sourcePath, commitMsg, nil
}
