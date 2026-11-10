package gitcode

import (
	"context"
	"errors"
	"fmt"
	"iter"
)

// maxPaginatePages bounds the page walk: per_page=100 × 50 pages is far
// beyond any real list, and a server that keeps returning full pages
// without the data advancing (broken cursor, ignored pagination params,
// cached response) used to walk forever — 125k requests in a 300ms repro —
// with only the caller's ctx as a brake (issue #5). The value aligns with
// the sister repo go-gongfeng's maxPaginatePages.
const maxPaginatePages = 50

// ErrTooManyPages is yielded by Paginate when the page walk hits
// maxPaginatePages and the last page was still full — the iteration was
// capped, not completed, and the caller is looking at a truncated list.
// Distinguish it with errors.Is; treating it as fatal or as "good enough"
// is the caller's call.
var ErrTooManyPages = errors.New("gitcode: paginate exceeded the page cap; results are truncated")

// PageResult represents a single page of results.
type PageResult[T any] struct {
	Items []T
	Page  int
}

// Paginate creates an iterator that automatically fetches all pages of results.
// Usage:
//
//	for items, err := range client.Paginate(ctx, func(opts ListOptions) ([]*Issue, error) {
//	    return client.ListIssues(ctx, owner, repo, ListIssuesOptions{ListOptions: opts})
//	}) {
//	    if err != nil { return err }
//	    for _, issue := range items {
//	        // process issue
//	    }
//	}
//
// The walk is bounded: at most maxPaginatePages pages are fetched. When the
// cap is hit on a still-full page the iterator yields ErrTooManyPages
// instead of ending silently, so a truncated result set cannot masquerade
// as a complete one (issue #5).
func Paginate[T any](ctx context.Context, fetch func(ListOptions) ([]T, error)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		perPage := 100 // max per page for efficiency

		for page := 1; page <= maxPaginatePages; page++ {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			items, err := fetch(ListOptions{Page: page, PerPage: perPage})
			if err != nil {
				yield(nil, err)
				return
			}

			if len(items) == 0 {
				return
			}

			if !yield(items, nil) {
				return
			}

			// If we got fewer items than requested, we've reached the end
			if len(items) < perPage {
				return
			}
		}

		// The cap is reached and the last page was still full: the caller
		// must know the list was truncated.
		yield(nil, fmt.Errorf("%w after %d pages of %d items", ErrTooManyPages, maxPaginatePages, perPage))
	}
}

// CollectAll collects all pages into a single slice.
func CollectAll[T any](ctx context.Context, fetch func(ListOptions) ([]T, error)) ([]T, error) {
	var all []T
	for items, err := range Paginate(ctx, fetch) {
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}
