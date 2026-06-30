package cmd

import (
	"encoding/json"
	"fmt"

	extv1 "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
)

// loginToken holds the fields extracted from a basic/OAuth login response,
// normalizing v3 Norman tokens and ext.cattle.io/v1 Tokens to a single type.
type loginToken struct {
	BearerToken string
	ExpiresAt   string
	UserID      string
}

// parseLoginResponse detects whether body is a v3 Norman token or an
// ext.cattle.io/v1 Token by inspecting apiVersion + kind, and returns a loginToken.
func parseLoginResponse(body []byte) (loginToken, error) {
	var hdr struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
	}
	if err := json.Unmarshal(body, &hdr); err != nil {
		return loginToken{}, fmt.Errorf("error unmarshaling login response: %w", err)
	}

	if hdr.APIVersion == "ext.cattle.io/v1" && hdr.Kind == "Token" {
		var t extv1.Token
		if err := json.Unmarshal(body, &t); err != nil {
			return loginToken{}, fmt.Errorf("error unmarshaling ext token response: %w", err)
		}
		return loginToken{
			BearerToken: t.Status.BearerToken,
			ExpiresAt:   t.Status.ExpiresAt,
			UserID:      t.Spec.UserID,
		}, nil
	}

	var t struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expiresAt"`
		UserID    string `json:"userId"`
	}
	if err := json.Unmarshal(body, &t); err != nil {
		return loginToken{}, fmt.Errorf("error unmarshaling v3 token response: %w", err)
	}
	return loginToken{
		BearerToken: t.Token,
		ExpiresAt:   t.ExpiresAt,
		UserID:      t.UserID,
	}, nil
}
