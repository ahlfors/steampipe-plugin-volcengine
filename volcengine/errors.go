package volcengine

import (
	"context"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// isNotFoundError returns an ErrorPredicateWithContext for Volcengine API calls
func isNotFoundError(notFoundErrors []string) plugin.ErrorPredicateWithContext {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData, err error) bool {
		config := GetConfig(d.Connection)

		allErrors := append(notFoundErrors, config.IgnoreErrorCodes...)

		for _, pattern := range allErrors {
			if strings.Contains(err.Error(), pattern) {
				return true
			}
		}
		return false
	}
}

// shouldIgnoreErrorPluginDefault returns the default error predicate for the plugin
func shouldIgnoreErrorPluginDefault() plugin.ErrorPredicateWithContext {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData, err error) bool {
		config := GetConfig(d.Connection)

		for _, pattern := range config.IgnoreErrorCodes {
			if strings.Contains(err.Error(), pattern) {
				return true
			}
		}
		return false
	}
}
