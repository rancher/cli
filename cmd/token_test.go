package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLoginResponse_V3(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	body := []byte(`{
		"type": "token",
		"token": "token-abc:secretxyz",
		"expiresAt": "` + expiresAt + `",
		"userId": "user-123"
	}`)

	got, err := parseLoginResponse(body)

	require.NoError(t, err)
	assert.Equal(t, "token-abc:secretxyz", got.BearerToken)
	assert.Equal(t, expiresAt, got.ExpiresAt)
	assert.Equal(t, "user-123", got.UserID)
}

func TestParseLoginResponse_Ext(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	body := []byte(`{
		"apiVersion": "ext.cattle.io/v1",
		"kind": "Token",
		"metadata": {"name": "token-def"},
		"spec": { "userID": "user-456" },
		"status": {
			"bearerToken": "ext/token-def:secretabc",
			"expiresAt": "` + expiresAt + `"
		}
	}`)

	got, err := parseLoginResponse(body)

	require.NoError(t, err)
	assert.Equal(t, "ext/token-def:secretabc", got.BearerToken)
	assert.Equal(t, expiresAt, got.ExpiresAt)
	assert.Equal(t, "user-456", got.UserID)
}

func TestParseLoginResponse_LooksLikeExtButWrongKind(t *testing.T) {
	t.Parallel()

	// Stray apiVersion field on a v3 response must not route to ext parser.
	body := []byte(`{
		"apiVersion": "management.cattle.io/v3",
		"type": "token",
		"token": "tok:sec",
		"expiresAt": "",
		"userId": "u"
	}`)

	got, err := parseLoginResponse(body)

	require.NoError(t, err)
	assert.Equal(t, "tok:sec", got.BearerToken)
	assert.Equal(t, "u", got.UserID)
}

func TestParseLoginResponse_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := parseLoginResponse([]byte("not json"))

	require.Error(t, err)
	assert.ErrorContains(t, err, "error unmarshaling login response")
}

func TestParseLoginResponse_ExtInvalidJSON(t *testing.T) {
	t.Parallel()

	body := []byte(`{"apiVersion": "ext.cattle.io/v1", "kind": "Token", "spec": "not-an-object"}`)

	_, err := parseLoginResponse(body)

	require.Error(t, err)
	assert.ErrorContains(t, err, "error unmarshaling ext token response")
}
