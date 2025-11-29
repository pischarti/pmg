package utils

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	cf "github.com/cloudflare/cloudflare-go"
)

func TestDecorateAuthError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if err := DecorateAuthError(nil); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("request error 6003", func(t *testing.T) {
		reqErr := cf.NewRequestError(&cf.Error{ErrorCodes: []int{6003}})
		err := DecorateAuthError(reqErr)
		if err == nil || err == reqErr {
			t.Fatalf("expected wrapped error, got %v", err)
		}
		if got := err.Error(); got == reqErr.Error() || !containsAll(got, []string{"6003", "invalid request headers"}) {
			t.Fatalf("expected guidance in error message, got %s", got)
		}
	})

	t.Run("api error 6003", func(t *testing.T) {
		apiErr := &cf.Error{ErrorCodes: []int{6003}}
		err := DecorateAuthError(fmt.Errorf("wrap: %w", apiErr))
		if err == nil || !containsAll(err.Error(), []string{"6003", "CLOUDFLARE_API_TOKEN"}) {
			t.Fatalf("expected hint, got %v", err)
		}
	})

	t.Run("other error passthrough", func(t *testing.T) {
		other := errors.New("boom")
		if err := DecorateAuthError(other); err != other {
			t.Fatalf("expected passthrough, got %v", err)
		}
	})
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}

func TestIsCloudflareError(t *testing.T) {
	reqErr := cf.NewRequestError(&cf.Error{ErrorCodes: []int{123}})
	if !IsCloudflareError(reqErr, 123) {
		t.Fatalf("expected code 123 to match")
	}
	if IsCloudflareError(reqErr, 999) {
		t.Fatalf("did not expect code 999 to match")
	}

	other := errors.New("boom")
	if IsCloudflareError(other, 123) {
		t.Fatalf("unexpected match for non-cloudflare error")
	}
}
