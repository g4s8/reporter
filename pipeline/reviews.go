package pipeline

import (
	"sync"

	"github.com/g4s8/reporter"
	"github.com/google/go-github/v32/github"
)

func MergeReviews(src []<-chan *reporter.PullWithReview) <-chan *reporter.PullWithReview {
	var wg sync.WaitGroup
	out := make(chan *reporter.PullWithReview)

	output := func(c <-chan *reporter.PullWithReview) {
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

func PullsWithReview(ctx *reporter.Context, pulls <-chan *reporter.PullResult,
	errs chan<- error) <-chan *reporter.PullWithReview {
	out := make(chan *reporter.PullWithReview)
	go func() {
		defer close(out)
		for pr := range pulls {
			opts := &github.ListOptions{}
			for {
				rev, rsp, err := ctx.Client.PullRequests.ListReviews(ctx,
					pr.Repo.GetOwner().GetLogin(), pr.Repo.GetName(),
					pr.Pull.GetNumber(), opts)
				if err != nil {
					errs <- err
					return
				}
				select {
				case out <- &reporter.PullWithReview{pr.Repo, pr.Pull, rev}:
					// nothing
				case <-ctx.Context.Done():
					errs <- ctx.Context.Err()
					return
				}
				if opts.Page < rsp.NextPage {
					opts.Page = rsp.NextPage
				} else {
					break
				}
			}
		}
	}()
	return out
}
