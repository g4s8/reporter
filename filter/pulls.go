package filter

import (
	"time"

	"github.com/g4s8/reporter"
)

type PullPredicate func(*reporter.PullResult) bool

var AllPulls = func(pull *reporter.PullResult) bool {
	return true
}

func (pred PullPredicate) And(other PullPredicate) PullPredicate {
	return func(pull *reporter.PullResult) bool {
		return pred(pull) && other(pull)
	}
}

func PullMergedAfter(t time.Time) PullPredicate {
	return func(pull *reporter.PullResult) bool {
		return pull.Pull.GetMergedAt().After(t)
	}
}

func PullFilter(src <-chan *reporter.PullResult,
	pred PullPredicate) <-chan *reporter.PullResult {
	out := make(chan *reporter.PullResult)
	go func() {
		defer close(out)
		for item := range src {
			if pred(item) {
				out <- item
			}
		}
	}()
	return out
}
