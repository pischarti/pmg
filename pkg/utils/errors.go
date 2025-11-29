package utils

import (
	"errors"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go"
)

// DecorateAuthError enriches Cloudflare client errors with actionable guidance for authentication issues.
func DecorateAuthError(err error) error {
	if err == nil {
		return nil
	}

	if isCloudflareErrorWithCode(err, 6003) {
		return fmt.Errorf("%w (cloudflare error 6003: invalid request headers); ensure you are using either a valid API token (CLOUDFLARE_API_TOKEN) or an API key plus email (CLOUDFLARE_API_KEY and CLOUDFLARE_API_EMAIL)", err)
	}

	return err
}

// IsCloudflareError reports whether the error returned by the Cloudflare SDK includes the given internal error code.
func IsCloudflareError(err error, code int) bool {
	return isCloudflareErrorWithCode(err, code)
}

func isCloudflareErrorWithCode(err error, code int) bool {
	var reqErr cf.RequestError
	if errors.As(err, &reqErr) && reqErr.InternalErrorCodeIs(code) {
		return true
	}

	var apiErr *cf.Error
	if errors.As(err, &apiErr) && apiErr.InternalErrorCodeIs(code) {
		return true
	}

	return false
}
