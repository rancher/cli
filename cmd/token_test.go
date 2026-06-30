package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rancher/norman/clientbase"
	extv1 "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNotFound() error {
	return &clientbase.APIError{StatusCode: http.StatusNotFound, Status: "404 Not Found", Msg: "not found"}
}

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

func TestGetExtToken_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/apis/ext.cattle.io/v1/tokens/token-abc", r.URL.Path)
		assert.Equal(t, "Bearer ext/token-abc:secret", r.Header.Get("Authorization"))
		fmt.Fprint(w, `{
			"apiVersion": "ext.cattle.io/v1",
			"kind": "Token",
			"metadata": {"name": "token-abc"},
			"spec": {"userID": "user-789"},
			"status": {"expired": false, "expiresAt": "2099-01-01T00:00:00Z"}
		}`)
	}))
	t.Cleanup(server.Close)

	token, err := getExtToken(context.Background(), "token-abc", server.URL, "ext/token-abc:secret", server.Client())

	require.NoError(t, err)
	assert.Equal(t, "user-789", token.Spec.UserID)
	assert.False(t, token.Status.Expired)
}

func TestGetExtToken_StripsExtPrefix(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Caller passed "ext/token-abc" as id; URL must not contain the slash-prefix.
		assert.Equal(t, "/apis/ext.cattle.io/v1/tokens/token-abc", r.URL.Path)
		fmt.Fprint(w, `{"apiVersion":"ext.cattle.io/v1","kind":"Token","spec":{"userID":"u"},"status":{"expired":false}}`)
	}))
	t.Cleanup(server.Close)

	_, err := getExtToken(context.Background(), "ext/token-abc", server.URL, "ext/token-abc:secret", server.Client())
	require.NoError(t, err)
}

func TestGetExtToken_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	_, err := getExtToken(context.Background(), "token-abc", server.URL, "token-abc:secret", server.Client())

	require.Error(t, err)
	assert.True(t, clientbase.IsNotFound(err), "expected clientbase.IsNotFound to be true")
}

func TestGetExtToken_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	_, err := getExtToken(context.Background(), "token-abc", server.URL, "token-abc:secret", server.Client())

	require.Error(t, err)
	assert.ErrorContains(t, err, "500")
}

func TestValidateToken_V3Valid(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) {
		return &managementClient.Token{Expired: false}, nil
	}
	ext := func(_ context.Context, _ string) (*extv1.Token, error) {
		t.Fatal("ext getter must not be called when v3 succeeds")
		return nil, nil
	}

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestValidateToken_V3Expired(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) {
		return &managementClient.Token{Expired: true}, nil
	}
	ext := func(_ context.Context, _ string) (*extv1.Token, error) {
		t.Fatal("ext getter must not be called when v3 succeeds")
		return nil, nil
	}

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateToken_V3NotFound_ExtValid(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, id string) (*extv1.Token, error) {
		return &extv1.Token{Status: extv1.TokenStatus{Expired: false}}, nil
	}

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestValidateToken_V3NotFound_ExtExpired(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, id string) (*extv1.Token, error) {
		return &extv1.Token{Status: extv1.TokenStatus{Expired: true}}, nil
	}

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateToken_V3NotFound_ExtNotFound(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, id string) (*extv1.Token, error) { return nil, newNotFound() }

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateToken_ExtPrefixSkipsV3(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) {
		t.Fatal("v3 getter must not be called for ext-prefixed token id")
		return nil, nil
	}
	ext := func(_ context.Context, id string) (*extv1.Token, error) {
		assert.Equal(t, "token-abc", id, "prefix must be stripped before ext lookup")
		return &extv1.Token{Status: extv1.TokenStatus{Expired: false}}, nil
	}

	ok, err := validateToken(context.Background(), "ext/token-abc", v3, ext)

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestValidateToken_V3NotFound_ExtUnauthorized(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, _ string) (*extv1.Token, error) {
		return nil, &clientbase.APIError{StatusCode: http.StatusUnauthorized, Status: "401 Unauthorized", Msg: "unauthorized"}
	}

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateToken_V3NotFound_ExtForbidden(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, _ string) (*extv1.Token, error) {
		return nil, &clientbase.APIError{StatusCode: http.StatusForbidden, Status: "403 Forbidden", Msg: "forbidden"}
	}

	ok, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateToken_V3NotFound_ExtOtherError(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, _ string) (*extv1.Token, error) {
		return nil, &clientbase.APIError{StatusCode: http.StatusInternalServerError, Status: "500", Msg: "boom"}
	}

	_, err := validateToken(context.Background(), "token-abc", v3, ext)

	require.Error(t, err)
}

func TestGetTokenUserID_V3(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) {
		return &managementClient.Token{UserID: "user-v3"}, nil
	}
	ext := func(_ context.Context, _ string) (*extv1.Token, error) {
		t.Fatal("ext getter must not be called when v3 succeeds")
		return nil, nil
	}

	uid, err := getTokenUserID(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.Equal(t, "user-v3", uid)
}

func TestGetTokenUserID_ExtFallback(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) { return nil, newNotFound() }
	ext := func(_ context.Context, id string) (*extv1.Token, error) {
		tok := &extv1.Token{}
		tok.Spec.UserID = "user-ext"
		return tok, nil
	}

	uid, err := getTokenUserID(context.Background(), "token-abc", v3, ext)

	require.NoError(t, err)
	assert.Equal(t, "user-ext", uid)
}

func TestGetTokenUserID_ExtPrefixSkipsV3(t *testing.T) {
	t.Parallel()

	v3 := func(id string) (*managementClient.Token, error) {
		t.Fatal("v3 getter must not be called for ext-prefixed token id")
		return nil, nil
	}
	ext := func(_ context.Context, id string) (*extv1.Token, error) {
		assert.Equal(t, "token-abc", id, "prefix must be stripped before ext lookup")
		tok := &extv1.Token{}
		tok.Spec.UserID = "user-ext"
		return tok, nil
	}

	uid, err := getTokenUserID(context.Background(), "ext/token-abc", v3, ext)

	require.NoError(t, err)
	assert.Equal(t, "user-ext", uid)
}
