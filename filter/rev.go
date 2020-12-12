package filter

import (
	"github.com/g4s8/reporter"
)

type ReviewPredicate func(*reporter.PullWithReview) bool

var AllReviews = func(pull *reporter.PullWithReview) bool {
	return true
}

func (pred ReviewPredicate) And(other ReviewPredicate) ReviewPredicate {
	return func(pull *reporter.PullWithReview) bool {
		return pred(pull) && other(pull)
	}
}

func RequestedChanges(min int) ReviewPredicate {
	return func(pull *reporter.PullWithReview) bool {
		var req int
		for _, r := range pull.Review {
			switch r.GetState() {
			case "CHANGES_REQUESTED":
				req++
				if req >= min {
					return true
				}
			}
		}
		return false
	}
}

var Approved = func(pull *reporter.PullWithReview) bool {
	for _, r := range pull.Review {
		switch r.GetState() {
		case "APPROVED":
			return true
		}
	}
	return false
}

func ReviewFilter(src <-chan *reporter.PullWithReview,
	pred ReviewPredicate) <-chan *reporter.PullWithReview {
	out := make(chan *reporter.PullWithReview)
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
