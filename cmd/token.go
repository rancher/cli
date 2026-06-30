package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rancher/norman/clientbase"
	extv1 "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
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

// tokenByIDFunc looks up a v3 Norman token by id (e.g. managementClient.TokenOperations.ByID).
type tokenByIDFunc func(id string) (*managementClient.Token, error)

// extTokenGetterFunc retrieves an ext.cattle.io/v1 Token by name.
type extTokenGetterFunc func(ctx context.Context, id string) (*extv1.Token, error)

// extTokenIDPrefix is the prefix Rancher prepends to ext-issued token bearer strings
// ("ext/<name>:<value>" per pkg/ext/stores/tokens/tokens.go:783 in rancher/rancher).
// When the CLI's stored access key carries this prefix, the id is an ext token and v3
// lookups must be skipped because they will always 404.
const extTokenIDPrefix = "ext/"

// getExtToken fetches a token from the ext.cattle.io/v1 API.
// baseURL is the Rancher server root without the trailing /v3 (use serverConfig.EnvironmentURL()).
// Returns a *clientbase.APIError with StatusCode 404 when the token is not found.
//
// IMPORTANT: do not wrap the returned *clientbase.APIError with fmt.Errorf("%w", ...) —
// clientbase.IsNotFound uses a direct type assertion (err.(*APIError)), not errors.As,
// so any wrap will make IsNotFound return false and break the v3-to-ext fallback contract.
//
// Strips an "ext/" prefix from id defensively; callers should already strip but a double strip is harmless.
func getExtToken(ctx context.Context, id, baseURL, bearerToken string, client *http.Client) (*extv1.Token, error) {
	id = strings.TrimPrefix(id, extTokenIDPrefix)
	u := strings.TrimRight(baseURL, "/") + "/apis/ext.cattle.io/v1/tokens/" + id

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating ext token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")

	resp, body, err := doRequest(client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// doRequest has already drained and closed resp.Body, so build the APIError
		// directly rather than calling clientbase.NewAPIError which would read a closed body.
		return nil, &clientbase.APIError{
			StatusCode: resp.StatusCode,
			URL:        u,
			Status:     resp.Status,
			Msg:        fmt.Sprintf("Bad response statusCode [%d]. Status [%s]. URL [%s]", resp.StatusCode, resp.Status, u),
			Body:       string(body),
		}
	}

	var t extv1.Token
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("error unmarshaling ext token: %w", err)
	}
	return &t, nil
}

// getTokenUserID returns the user id associated with the given token id.
// It tries the v3 Management API first and falls back to ext.cattle.io/v1.
// Ids prefixed with "ext/" skip the v3 attempt entirely. The prefix is stripped
// before the ext lookup so callers receive a clean token name.
func getTokenUserID(ctx context.Context, tokenID string, v3ByID tokenByIDFunc, extGetter extTokenGetterFunc) (string, error) {
	if !strings.HasPrefix(tokenID, extTokenIDPrefix) {
		token, err := v3ByID(tokenID)
		if err == nil {
			return token.UserID, nil
		}
		if !clientbase.IsNotFound(err) {
			return "", err
		}
	}

	extToken, err := extGetter(ctx, strings.TrimPrefix(tokenID, extTokenIDPrefix))
	if err != nil {
		return "", fmt.Errorf("error resolving user id for token %q: %w", tokenID, err)
	}
	return extToken.Spec.UserID, nil
}

// validateToken reports whether the token with the given id exists and is not expired.
// It tries the v3 Management API first and falls back to ext.cattle.io/v1.
// Ids prefixed with "ext/" skip the v3 attempt entirely. The prefix is stripped
// before the ext lookup so callers receive a clean token name.
func validateToken(ctx context.Context, tokenID string, v3ByID tokenByIDFunc, extGetter extTokenGetterFunc) (bool, error) {
	if !strings.HasPrefix(tokenID, extTokenIDPrefix) {
		token, err := v3ByID(tokenID)
		if err == nil {
			return !token.Expired, nil
		}
		if !clientbase.IsNotFound(err) {
			return false, err
		}
	}

	extToken, err := extGetter(ctx, strings.TrimPrefix(tokenID, extTokenIDPrefix))
	if err != nil {
		// 404/401/403 from the ext API are all "could not determine validity":
		// the token isn't there, or the bearer format isn't accepted by this
		// server. Fall through to kubeconfig regeneration in either case rather
		// than surfacing an opaque error to the user.
		if clientbase.IsNotFound(err) || isUnauthorized(err) {
			return false, nil
		}
		return false, err
	}
	return !extToken.Status.Expired, nil
}

// isUnauthorized reports whether err is a *clientbase.APIError with status 401 or 403.
func isUnauthorized(err error) bool {
	var apiErr *clientbase.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden
	}
	return false
}
