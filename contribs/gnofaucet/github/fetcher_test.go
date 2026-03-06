package github

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/go-github/v74/github"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var noopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestFetch(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	githubClient := github.NewClient(nil)
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")
	githubGraphql := graphql.NewClient(server.URL+"/graphql", nil)
	ghImpl := NewGithubClientImpl(githubClient, githubGraphql)

	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	repos := map[string][]string{
		"gnolang": {"gno"},
	}

	fetcher := NewGHFetcher(ghImpl, rdb, repos, noopLogger, 1*time.Second)

	ctx := context.Background()

	err := fetcher.fetchHistory(ctx)
	require.NoError(t, err)
	pipe := rdb.Pipeline()
	out := fetcher.iterateEvents(ctx, pipe, "gnolang", "gno")
	require.True(t, out)
	_, err = pipe.Exec(ctx)
	require.NoError(t, err)

	keys, err := rdb.Keys(ctx, "*").Result()
	require.NoError(t, err)
	require.Len(t, keys, 134)

	require.Equal(t, "1", getResult(t, rdb, "issue:gfanton"))
	require.Equal(t, "21", getResult(t, rdb, "prr:moul"))
	require.Equal(t, "1", getResult(t, rdb, "pr:Davphla"))
	require.Equal(t, "84", getResult(t, rdb, "commit:jaekwon"))
	require.Equal(t, "2025-07-30T09:44:14Z", getResult(t, rdb, "lastFetch:gnolang:gno"))
}

func getResult(t *testing.T, rdb *redis.Client, key string) string {
	t.Helper()
	res, err := rdb.Get(context.Background(), key).Result()
	require.NoError(t, err)
	return res
}

func createTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle different endpoints
		switch {
		case strings.Contains(r.URL.Path, "/repos/gnolang/gno/issues"):
			handleIssuesRequest(t, w, r)
		case strings.Contains(r.URL.Path, "/repos/gnolang/gno/events"):
			handleEventsRequest(t, w, r)
		case r.URL.Path == "/graphql":
			handleGraphQLRequest(t, w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

func handleIssuesRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	// Load test data from file
	data, err := os.ReadFile(filepath.Join("testdata", "issues."+page+".json"))
	require.NoError(t, err)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleEventsRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	// Load test data from file
	data, err := os.ReadFile(filepath.Join("testdata", "events."+page+".json"))
	require.NoError(t, err)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleGraphQLRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	obj := &graphQLRequest{}
	err := json.NewDecoder(r.Body).Decode(obj)
	require.NoError(t, err)

	page := "1"
	if obj.Variables.Cursor == "Y3Vyc29yOnYyOpK5MjAyMi0wNC0yMFQxOTozODoxMSswMjowMM42gett" {
		page = "2"
	}

	// Load test data from file and fix the format
	data, err := os.ReadFile(filepath.Join("testdata", "pulls."+page+".json"))
	require.NoError(t, err)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type graphQLRequest struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
	Variables     struct {
		Cursor string `json:"cursor"`
		Name   string `json:"name"`
		Owner  string `json:"owner"`
	} `json:"variables"`
}
