package view

import (
	"fmt"
	"github.com/g4s8/reporter"
)

type PullsView interface {
	Draw(<-chan *reporter.PullWithReview) error
}

type PullsViewMode uint8

const (
	PullsViewStats PullsViewMode = iota
	PullsViewSummary
	PullsViewDetails
)

func NewPullsView(p Printer, mode PullsViewMode, links bool) (PullsView, error) {
	switch mode {
	case PullsViewSummary:
		return &pullsSummary{p, links}, nil
	case PullsViewStats:
		return &pullsStats{p, 0, make(map[string]int)},
			nil
	default:
		return nil, fmt.Errorf("unsupported mode %d", mode)
	}
}

type pullsSummary struct {
	Printer
	links bool
}

func (v *pullsSummary) Draw(pulls <-chan *reporter.PullWithReview) error {
	for pull := range pulls {
		txt := fmt.Sprintf("#%d \"%s\" (by @%s), rev(%d)",
			pull.Pull.GetNumber(),
			pull.Pull.GetTitle(), pull.Pull.GetUser().GetLogin(),
			len(pull.Review))
		if v.links {
			txt += " " + pull.Pull.GetHTMLURL()
		}
		v.Printer.Print(txt)
	}
	return nil
}

type pullsStats struct {
	Printer
	total  int
	byUser map[string]int
}

func (v *pullsStats) Draw(pulls <-chan *reporter.PullWithReview) error {
	for pull := range pulls {
		user := pull.Pull.User.GetLogin()
		v.total++
		var byUser int
		if u, ok := v.byUser[user]; ok {
			byUser = u
		}
		byUser++
		v.byUser[user] = byUser
	}
	v.Printer.Print(fmt.Sprintf("%d pull requests found:", v.total))
	for user, count := range v.byUser {
		v.Printer.Print(fmt.Sprintf(" - @%s - %d PRs", user, count))
	}
	return nil
}
