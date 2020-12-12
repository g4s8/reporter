package pipeline

import (
	"fmt"
	"strings"

	"github.com/g4s8/reporter"
	"github.com/google/go-github/v32/github"
)

type errBadRepoName struct {
	name string
}

func (e *errBadRepoName) Error() string {
	return fmt.Sprintf("bad repository name: `%s`", e.name)
}

// ReposByName fetches repository for each name in source channel
func ReposByNames(ctx *reporter.Context, names <-chan string,
	errs chan<- error) <-chan *github.Repository {
	out := make(chan *github.Repository)
	go func() {
		defer close(out)
		for name := range names {
			parts := strings.Split(name, "/")
			if len(parts) == 1 {
				repos, _, err := ctx.Client.Repositories.ListByOrg(ctx,
					parts[0], nil)
				if err != nil {
					errs <- err
					return
				}
				for _, r := range repos {
					select {
					case out <- r:
					case <-ctx.Context.Done():
						errs <- ctx.Context.Err()
						return
					}
				}
			} else if len(parts) == 2 {
				r, _, err := ctx.Client.Repositories.Get(ctx,
					parts[0], parts[1])
				if err != nil {
					errs <- err
					return
				}
				select {
				case out <- r:
				case <-ctx.Context.Done():
					errs <- ctx.Context.Err()
					return
				}
			} else {
				errs <- &errBadRepoName{name}
				return
			}
		}
	}()
	return out
}
