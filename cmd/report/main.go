package main

import (
	"fmt"
	"time"
	// "github.com/caarlos0/spin"
	"context"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/g4s8/reporter"
	"github.com/g4s8/reporter/filter"
	"github.com/g4s8/reporter/pipeline"
	"github.com/g4s8/reporter/view"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:                 "reporter",
		Usage:                "GitHub report generator and statistics aggregator",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "token",
				Usage: "GitHub API token with",
			},
			&cli.StringFlag{
				Name:  "verbose",
				Usage: "Verbose output",
			},
		},
		Before: setup,
		Commands: []*cli.Command{
			{
				Name:    "pulls",
				Aliases: []string{"p"},
				Usage:   "Show pull requests statistics",
				Action:  cmdPulls,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "view",
						Value: "summary",
						Usage: "Report view: `[stats|summary|details]`, `summary` by default",
					},
					&cli.StringFlag{
						Name:  "since",
						Usage: "start date to filter PRs",
					},
					&cli.BoolFlag{
						Name:  "require-review",
						Usage: "Filter pull requests only with review changes fixed and approved",
					},
					&cli.BoolFlag{
						Name:  "include-links",
						Usage: "Add PR links to view if supported",
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var ctx *reporter.Context

func setup(app *cli.Context) error {
	tkn, err := token(app)
	if err != nil {
		return err
	}
	canc, _ := context.WithTimeout(app.Context, time.Minute*3)
	ctx = reporter.New(canc, tkn)
	return nil
}

func token(app *cli.Context) (string, error) {
	token := app.String("token")
	if token != "" {
		return token, nil
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t, nil
	}
	file := os.Getenv("HOME") + "/.config/reporter/github_token.txt"
	if token, err := ioutil.ReadFile(file); err == nil {
		return string(token), nil
	}
	return "", fmt.Errorf("GitHub token neither given as a flag, nor found in env, not in %s", file)
}

type fmtPrinter struct {
	mux *sync.Mutex
}

func (p *fmtPrinter) Print(text string) {
	p.mux.Lock()
	fmt.Println(text)
	p.mux.Unlock()
}

func cmdPulls(app *cli.Context) error {
	args := app.Args()
	if !args.Present() {
		return fmt.Errorf("requires at least one repostiry argument")
	}
	printer := &fmtPrinter{new(sync.Mutex)}
	var mode view.PullsViewMode
	switch app.String("view") {
	case "summary":
		mode = view.PullsViewSummary
	case "stats":
		mode = view.PullsViewStats
	case "details":
		mode = view.PullsViewDetails
	default:
		return fmt.Errorf("unsupported view mode `%s`", app.String("view"))
	}

	view, err := view.NewPullsView(printer, mode, app.Bool("include-links"))
	if err != nil {
		return err
	}
	since := app.String("since")
	var pullPred filter.PullPredicate = filter.AllPulls
	var revPred filter.ReviewPredicate = filter.AllReviews
	if since != "" {
		after, err := time.Parse("2006-01-02", since)
		if err != nil {
			return err
		}
		pullPred = pullPred.And(filter.PullMergedAfter(after))
	}
	if app.Bool("require-review") {
		revPred = revPred.And(filter.Approved).And(filter.RequestedChanges(1))
	}
	errs := make(chan error)
	nc := make(chan string)
	repochan := pipeline.ReposByNames(ctx, nc, errs)
	var prchans [2]<-chan *reporter.PullResult
	for i := 0; i < len(prchans); i++ {
		prchans[i] = pipeline.ReposToPulls(ctx, repochan, errs)
	}
	prchan := filter.PullFilter(pipeline.MergePulls(prchans[:]), pullPred)
	var revchans [6]<-chan *reporter.PullWithReview
	for i := 0; i < len(revchans); i++ {
		revchans[i] = pipeline.PullsWithReview(ctx, prchan, errs)
	}
	revchan := filter.ReviewFilter(pipeline.MergeReviews(revchans[:]), revPred)

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := view.Draw(revchan); err != nil {
			errs <- err
		}
	}()

	for _, name := range app.Args().Slice() {
		nc <- name
	}
	close(nc)
	select {
	case <-done:
		return nil
	case err := <-errs:
		return err
	case <-ctx.Context.Done():
		return ctx.Context.Err()
	}
}
