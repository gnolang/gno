// Package indexer is a minimal client for a tx-indexer GraphQL endpoint.
// It knows nothing about gnoweb, which is what keeps it removable.
package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// ErrUnavailable is returned while the breaker is open, without a request.
var ErrUnavailable = errors.New("indexer unavailable, not retried yet")

// ErrNotFound describes the question, not the indexer's health, so it never
// counts against the breaker.
var ErrNotFound = errors.New("not found")

// ErrResponseTooLarge is refused before decoding, and counts against the
// breaker: it describes the endpoint.
var ErrResponseTooLarge = errors.New("indexer response too large")

// ErrTooLarge reports the server-side element cap. The response is partial,
// not empty: the resolver returns the rows it walked before stopping.
var ErrTooLarge = errors.New("indexer result set hit the element cap")

// The tx-indexer reports a miss and a cap as GraphQL error text, with no code
// to switch on.
const (
	notFoundMessage = "item not found in storage"
	tooLargeMessage = "max elements per query"
)

const (
	// defaultTimeout bounds a single request: a page render must not wait on
	// the indexer longer than a reader waits on the page.
	defaultTimeout = 4 * time.Second

	// maxResponseSize mirrors maxRPCResponseSize. The indexer is off-chain
	// and operator-configured, so it is trusted no further than the node.
	maxResponseSize = 8 << 20

	// Without a breaker, a dead indexer adds the full timeout to every
	// request that consults it.
	breakerThreshold = 3
	breakerCooldown  = 30 * time.Second

	// maxConcurrent mirrors the RPC client's semaphore: each in-flight
	// request can hold maxResponseSize plus its decoded form.
	maxConcurrent = 16

	// http.DefaultTransport allows 2, which turns a burst of searches into a
	// TLS handshake each.
	maxIdleConnsPerHost = 32
)

// Client is a tx-indexer GraphQL client. Callers hold it behind a nilable
// interface; that nil check is the feature switch.
type Client struct {
	url string
	// display is url with its credentials removed: the only form of the
	// endpoint that may reach a page, a JSON response or a log line.
	display string
	// token, when set, is sent as a bearer credential. Indexers are usually
	// public, so most deployments leave it empty.
	token string
	http  *http.Client

	// slots bounds in-flight requests. Buffered, never closed.
	slots chan struct{}

	mu        sync.Mutex
	failures  int
	skipUntil time.Time

	// Chain tip, cached for heightTTL and refreshed on request, never by a
	// ticker. tipMu guards the value only: holding it across the fetch would
	// serialize every search behind one slow-but-healthy indexer, which the
	// breaker never catches because nothing fails.
	tipMu    sync.Mutex
	tip      int
	tipAt    time.Time
	tipGroup singleflight.Group
}

// New returns a Client for a tx-indexer GraphQL endpoint, e.g.
// https://indexer.gno.land/graphql/query. An empty token means no
// Authorization header is sent.
//
// The URL may carry credentials, as userinfo or a query-string key; they are
// sent to the indexer and never shown: URL reports the redacted form.
func New(rawURL, token string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = maxIdleConnsPerHost

	return &Client{
		url:     rawURL,
		display: Redact(rawURL),
		token:   token,
		slots:   make(chan struct{}, maxConcurrent),
		http: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
			// A redirect would let the indexer aim gnoweb at a host nobody
			// configured, an internal address included.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// URL reports the configured endpoint, for display and health reporting,
// as scheme and host only. See Redact.
func (c *Client) URL() string { return c.display }

// ValidateURL rejects an endpoint the client could never reach, so a typo
// fails at startup rather than behind a log line calling the indexer enabled.
// `localhost:8546/graphql` parses, with "localhost" as its scheme.
func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid indexer URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid indexer URL: scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("invalid indexer URL: no host")
	}
	return nil
}

// Redact reduces an endpoint to its scheme and host. Userinfo and a query
// string are where an operator puts an indexer credential, and the endpoint
// is shown to every anonymous reader; the path goes too, since a key can sit
// there as well.
func Redact(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "(indexer)"
	}
	return u.Scheme + "://" + u.Host
}

// Query executes a GraphQL document and decodes `data` into out.
//
// The query is a raw string so several fields can ride in one document.
// Values are interpolated through gqlString — deliberately the only escaping
// mechanism here.
func (c *Client) Query(ctx context.Context, query string, out any) error {
	if c.breakerOpen() {
		return ErrUnavailable
	}

	if err := c.acquire(ctx); err != nil {
		return err
	}
	err := c.do(ctx, query, out)
	c.release()

	// A miss, a cap and a caller giving up say nothing about the indexer's
	// health. The client's own timeout does, and is the signal the breaker
	// acts on. A caller's deadline is the caller's budget, not the
	// indexer's: the omnibar gives up at 3s, below the client's 4s, and
	// three readers typing at once would otherwise close the indexer to
	// everyone for the cooldown.
	switch {
	case err == nil:
		c.record(nil)
	case ctx.Err() != nil,
		errors.Is(err, ErrNotFound),
		errors.Is(err, ErrTooLarge):
	default:
		c.record(err)
	}
	return err
}

// acquire takes a slot or gives up with the caller's context.
func (c *Client) acquire(ctx context.Context) error {
	select {
	case c.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) release() { <-c.slots }

func (c *Client) breakerOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Now().Before(c.skipUntil)
}

func (c *Client) record(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		c.failures, c.skipUntil = 0, time.Time{}
		return
	}
	if c.failures++; c.failures >= breakerThreshold {
		c.skipUntil = time.Now().Add(breakerCooldown)
	}
}

type gqlRequest struct {
	Query string `json:"query"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) do(ctx context.Context, query string, out any) error {
	body, err := json.Marshal(gqlRequest{Query: query})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// net/http quotes the request URL in its error, masking a password
		// but not a query-string key. These errors are logged.
		var uerr *url.Error
		if errors.As(err, &uerr) {
			uerr.URL = c.display
		}
		return err
	}
	defer resp.Body.Close()

	// One byte over the cap, so a body ending exactly on it is not mistaken
	// for a truncated one.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return err
	}
	if len(raw) > maxResponseSize {
		return fmt.Errorf("%w: over %d bytes", ErrResponseTooLarge, maxResponseSize)
	}

	// Name the status: a 403 otherwise surfaces as "invalid character '<'".
	// The body is never quoted, only its size and type: these errors are
	// logged, and visitors see a fixed phrase.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("indexer returned %s (%d bytes of %q)",
			resp.Status, len(raw), resp.Header.Get("Content-Type"))
	}

	var gql gqlResponse
	if err := json.Unmarshal(raw, &gql); err != nil {
		return fmt.Errorf("decode response: %w (%d bytes of %q)",
			err, len(raw), resp.Header.Get("Content-Type"))
	}

	if len(gql.Errors) > 0 {
		return c.gqlError(gql, out)
	}
	if len(gql.Data) == 0 || string(gql.Data) == "null" {
		// GraphQL always answers `data`, `errors` or both. Neither is what
		// an -indexer-url pointing at the wrong endpoint returns, and taking
		// it as an empty answer read as "not found" and a tip of 0.
		return fmt.Errorf("indexer answered neither data nor errors (%d bytes of %q)",
			len(raw), resp.Header.Get("Content-Type"))
	}
	return json.Unmarshal(gql.Data, out)
}

// gqlError classifies a GraphQL-level error. Only the element cap carries
// usable data alongside it.
func (c *Client) gqlError(gql gqlResponse, out any) error {
	for _, e := range gql.Errors {
		if !strings.Contains(e.Message, tooLargeMessage) {
			continue
		}
		// `data` is optional here: an indexer may report the cap with no
		// partial rows. Losing the sentinel would make a capped answer a
		// breaker failure.
		if len(gql.Data) > 0 {
			if err := json.Unmarshal(gql.Data, out); err != nil {
				return fmt.Errorf("decode capped response: %w", err)
			}
		}
		return fmt.Errorf("%w: %s", ErrTooLarge, e.Message)
	}
	if strings.Contains(gql.Errors[0].Message, notFoundMessage) {
		return fmt.Errorf("%w: %s", ErrNotFound, gql.Errors[0].Message)
	}
	return fmt.Errorf("graphql error: %s", gql.Errors[0].Message)
}
