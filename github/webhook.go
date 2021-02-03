package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Parsing error message
var (
	ErrEventNotSpecifiedToParse	 = errors.New("No Event specified to parse")
	ErrInvalidHTTPMethod         = errors.New("Invalid HTTP Method")
	ErrMissingGithubEventHeader  = errors.New("Missing X-GitHub-Event Header")
	ErrMissingHubSignatureHeader = errors.New("Missing X-Hub-Signature Header")
	ErrEventNotFound             = errors.New("Event not defined to be parsed")
	ErrParsingPayload            = errors.New("Error parsing payload")
	ErrHMACVerificationFailed    = errors.New("HMAC verification failed")
)

type Event string

const (
	CheckRunEvent                            Event = "check_run"
	CheckSuiteEvent                          Event = "check_suite"
	CommitCommentEvent                       Event = "commit_comment"
	CreateEvent                              Event = "create"
	DeleteEvent                              Event = "delete"
	DeploymentEvent                          Event = "deployment"
	DeploymentStatusEvent                    Event = "deployment_status"
	ForkEvent                                Event = "fork"
	GollumEvent                              Event = "gollum"
	InstallationEvent                        Event = "installation"
	InstallationRepositoriesEvent            Event = "installation_repositories"
	IntegrationInstallationEvent             Event = "integration_installation"
	IntegrationInstallationRepositoriesEvent Event = "integration_installation_repositories"
	IssueCommentEvent                        Event = "issue_comment"
	IssuesEvent                              Event = "issues"
	LabelEvent                               Event = "label"
	MemberEvent                              Event = "member"
	MembershipEvent                          Event = "membership"
	MilestoneEvent                           Event = "milestone"
	MetaEvent                                Event = "meta"
	OrganizationEvent                        Event = "organization"
	OrgBlockEvent                            Event = "org_block"
	PageBuildEvent                           Event = "page_build"
	PingEvent                                Event = "ping"
	ProjectCardEvent                         Event = "project_card"
	ProjectColumnEvent                       Event = "project_column"
	ProjectEvent                             Event = "project"
	PublicEvent                              Event = "public"
	PullRequestEvent                         Event = "pull_request"
	PullRequestReviewEvent                   Event = "pull_request_review"
	PullRequestReviewCommentEvent            Event = "pull_request_review_comment"
	PushEvent                                Event = "push"
	ReleaseEvent                             Event = "release"
	RepositoryEvent                          Event = "repository"
	RepositoryVulnerabilityAlertEvent        Event = "repository_vulnerability_alert"
	SecurityAdvisoryEvent                    Event = "security_advisory"
	StatusEvent                              Event = "status"
	TeamEvent                                Event = "team"
	TeamAddEvent                             Event = "team_add"
	WatchEvent                               Event = "watch"
)

type EventSubtype string

// GitHub hook event subtypes
const (
	NoSubtype     EventSubtype = ""
	BranchSubtype EventSubtype = "branch"
	TagSubtype    EventSubtype = "tag"
	PullSubtype   EventSubtype = "pull"
	IssueSubtype  EventSubtype = "issues"
)

// Webhook instance contains all methods needed to process events
type Webhook struct {
	secret string
}

// Option is a configuration option for webhook
type Option func(*Webhook) error

// Option is a namespace var for configuration options
var Options = WebhookOptions{}

// Option is a namespace var for configuration option methods
type WebhookOptions struct{}


// Secret registers the GitHub secret
func (WebhookOptions) Secret(secret string) Option {
	return func(hook *Webhook) error {
		hook.secret = secret
		return nil
	}
}

// Create and return new Webhook instance denoted by the Provider(github) type
func New(options ...Option) (*Webhook, error) {
		hook := new(Webhook)
		for _, opt := range options {
				if err := opt(hook); err != nil {
						return nil, errors.New("Error applying Option")
				}
		}
		return hook, nil
}

