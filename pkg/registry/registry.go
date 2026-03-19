package registry

import (
	"context"
	"fmt"

	"github.com/containrrr/watchtower/pkg/registry/helpers"
	"github.com/containrrr/watchtower/pkg/retry"
	watchtowerTypes "github.com/containrrr/watchtower/pkg/types"
	ref "github.com/distribution/reference"
	"github.com/docker/docker/api/types/image"
	log "github.com/sirupsen/logrus"
)

// GetPullOptions creates a struct with all options needed for pulling images from a registry
func GetPullOptions(imageName string) (image.PullOptions, error) {
	// Enhanced logging for registry authentication
	authEntry := log.WithFields(log.Fields{
		"image_name": imageName,
		"module":     "registry",
		"operation":  "get_pull_options",
	})

	// Use retry logic for authentication
	config := retry.DefaultConfig()

	var auth string
	var err error

	authFn := func() error {
		auth, err = EncodedAuth(imageName)
		authEntry.Debugf("Processing image name: %s", imageName)
		return err
	}

	_, retryErr := retry.WithRetry(context.Background(), config, fmt.Sprintf("registry_auth_%s", imageName), authFn)
	if retryErr != nil {
		authEntry.WithError(retryErr).Error("Failed to get registry authentication")
		return image.PullOptions{}, retryErr
	}

	if auth == "" {
		authEntry.Debug("No authentication credentials required")
		return image.PullOptions{}, nil
	}

	authEntry.Debug("Authentication credentials retrieved successfully")

	// CREDENTIAL: Uncomment to log docker config auth
	// log.Tracef("Got auth value: %s", auth)

	return image.PullOptions{
		RegistryAuth:  auth,
		PrivilegeFunc: DefaultAuthHandler,
	}, nil
}

// DefaultAuthHandler will be invoked if an AuthConfig is rejected
// It could be used to return a new value for the "X-Registry-Auth" authentication header,
// but there's no point trying again with the same value as used in AuthConfig
func DefaultAuthHandler(ctx context.Context) (string, error) {
	log.WithFields(log.Fields{
		"module":    "registry",
		"operation": "auth_handler",
	}).Debug("Authentication request was rejected, trying again without authentication")
	return "", nil
}

// WarnOnAPIConsumption will return true if the registry is known-expected
// to respond well to HTTP HEAD in checking the container digest -- or if there
// are problems parsing the container hostname.
// Will return false if behavior for container is unknown.
func WarnOnAPIConsumption(container watchtowerTypes.Container) bool {

	normalizedRef, err := ref.ParseNormalizedNamed(container.ImageName())
	if err != nil {
		return true
	}

	containerHost, err := helpers.GetRegistryAddress(normalizedRef.Name())
	if err != nil {
		return true
	}

	if containerHost == helpers.DefaultRegistryHost || containerHost == "ghcr.io" {
		return true
	}

	return false
}
