package pipeline

import (
	"sync"

	"github.com/g4s8/reporter"
	"github.com/google/go-github/v32/github"
)

func ReposToPulls(ctx *reporter.Context, repos <-chan *github.Repository,
	errs chan<- error) <-chan *reporter.PullResult {
	out := make(chan *reporter.PullResult)
	go func() {
		defer close(out)
		for repo := range repos {
			opts := &github.PullRequestListOptions{
				State: "closed",
			}
			for {
				res, rsp, err := ctx.Client.PullRequests.List(ctx,
					repo.GetOwner().GetLogin(), repo.GetName(), opts)
				if err != nil {
					errs <- err
					return
				}
				for _, pr := range res {
					select {
					case out <- &reporter.PullResult{repo, pr}:
						continue
					case <-ctx.Context.Done():
						errs <- ctx.Context.Err()
						return
					}
				}
				if opts.Page < rsp.LastPage {
					opts.Page = rsp.NextPage
				} else {
					break
				}
			}

		}
	}()
	return out
}

func MergePulls(src []<-chan *reporter.PullResult) <-chan *reporter.PullResult {
	var wg sync.WaitGroup
	out := make(chan *reporter.PullResult)

	output := func(c <-chan *reporter.PullResult) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(src))
	for _, c := range src {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
