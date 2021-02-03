package webhook

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"errors"
	"os"
	"testing"
	"io"
	"reflect"

	"github.com/TRQ1/webhook-cevier-go/webhook"
	"github.com/TRQ1/webhook-reciver-go/github_payload"
	"github.com/stretchr/testify/require"
)


const (
	path = "/webhooks"
)

var hook *Webhook

func TestMain(m *testing.M) {

	// setup
	var err error
	hook, err= New(Options.Secret("IsWishesWereHorsesWedAllBeEatingSteak"))
	if err != nil {
			log.Fatal(err)
	}
	os.Exit(m.Run())
	// teardown
}

func newServer(handler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(path, handler)
	return httptest.NewServer(mux)
}

func TestBadRequests(t *testing.T) {

	assert := require.New(t)
	tests := []struct {
		name    string
		event   Event
		payload io.Reader
		headers http.Header
	}{
		{
			name:    "BadNoEventHeader",
			event:   CreateEvent,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{},
		},
		{
			name:    "UnsubscribedEvent",
			event:   CreateEvent,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Github-Event": []string{"noneexistant_event"},
			},
		},
		{
			name:    "BadBody",
			event:   CommitCommentEvent,
			payload: bytes.NewBuffer([]byte("")),
			headers: http.Header{
				"X-Github-Event":  []string{"commit_comment"},
				"X-Hub-Signature": []string{"sha1=156404ad5f721c53151147f3d3d302329f95a3ab"},
			},
		},
		{
			name:    "BadSignatureLength",
			event:   CommitCommentEvent,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Github-Event":  []string{"commit_comment"},
				"X-Hub-Signature": []string{""},
			},
		},
		{
			name:    "BadSignatureMatch",
			event:   CommitCommentEvent,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Github-Event":  []string{"commit_comment"},
				"X-Hub-Signature": []string{"sha1=111"},
			},
		},
	}
	for _, tt := range tests {
		tc := tt
		client := &http.Client{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var parseError error
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, parseError = hook.Parse(r, tc.event)
			})
			defer server.Close()
			req, err := http.NewRequest(http.MethodPost, server.URL+path, tc.payload)
			assert.NoError(err)
			req.Header = tc.headers
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			assert.NoError(err)
			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.Error(parseError)
		})
	}
}

func TestWebhook(t *testing.T) {
	assert := require.New(t)
	tests := []struct {
		name     string
		event    Event
		typ      interface{}
		filename string
		headers  http.Header
	}{
		{
			name:     "CheckRunEvent",
			event:    CheckRunEvent,
			typ:      CheckRunPayload{},
			filename: "../testdata/check-run.json",
			headers: http.Header{
				"X-Github-Event":  []string{"check_run"},
				"X-Hub-Signature": []string{"sha1=229f4920493b455398168cd86dc6b366064bdf3f"},
			},
		},
		{
			name:     "CheckSuiteEvent",
			event:    CheckSuiteEvent,
			typ:      CheckSuitePayload{},
			filename: "../testdata/check-suite.json",
			headers: http.Header{
				"X-Github-Event":  []string{"check_suite"},
				"X-Hub-Signature": []string{"sha1=250ad5a340f8d91e67dc5682342f3190fd2006a1"},
			},
		},
		{
			name:     "PullRequestEvent",
			event:    PullRequestEvent,
			typ:      PullRequestPayload{},
			filename: "../testdata/pull-request.json",
			headers: http.Header{
				"X-Github-Event":  []string{"pull_request"},
				"X-Hub-Signature": []string{"sha1=88972f972db301178aa13dafaf112d26416a15e6"},
			},
		},
		{
			name:     "PullRequestReviewEvent",
			event:    PullRequestReviewEvent,
			typ:      PullRequestReviewPayload{},
			filename: "../testdata/pull-request-review.json",
			headers: http.Header{
				"X-Github-Event":  []string{"pull_request_review"},
				"X-Hub-Signature": []string{"sha1=55345ce92be7849f97d39b9426b95261d4bd4465"},
			},
		},
		{
			name:     "PullRequestReviewCommentEvent",
			event:    PullRequestReviewCommentEvent,
			typ:      PullRequestReviewCommentPayload{},
			filename: "../testdata/ull-request-review-comment.json",
			headers: http.Header{
				"X-Github-Event":  []string{"pull_request_review_comment"},
				"X-Hub-Signature": []string{"sha1=a9ece15dbcbb85fa5f00a0bf409494af2cbc5b60"},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		client := &http.Client{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload, err := os.Open(tc.filename)
			assert.NoError(err)
			defer func() {
				_ = payload.Close()
			}()

			var parseError error
			var results interface{}
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				results, parseError = hook.Parse(r, tc.event)
			})
			defer server.Close()
			req, err := http.NewRequest(http.MethodPost, server.URL+path, payload)
			assert.NoError(err)
			req.Header = tc.headers
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			assert.NoError(err)
			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.NoError(parseError)
			assert.Equal(reflect.TypeOf(tc.typ), reflect.TypeOf(results))
		})
	}
}