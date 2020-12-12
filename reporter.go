package reporter

import (
	"context"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type Context struct {
	context.Context
	*github.Client
}

func New(ctx context.Context, tkn string) *Context {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tkn},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Context{Client: github.NewClient(tc), Context: ctx}
}

type PullResult struct {
	Repo *github.Repository
	Pull *github.PullRequest
}

type PullWithReview struct {
	Repo   *github.Repository
	Pull   *github.PullRequest
	Review []*github.PullRequestReview
}
