package gitcode

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// escapedPath marks an argument that is already percent-escaped and must be
// inserted into the path verbatim (multi-segment file paths escaped by
// pathSegs). pathf leaves values of this type untouched so they are not
// double-escaped.
type escapedPath string

// pathSeg escapes a single path segment (owner, repo, org, branch, sha…).
//
// Callers control these values and they must not be able to alter the shape
// of the request path: url.PathEscape turns "/" into %2F, "?" into %3F,
// "#" into %23 — a malicious or merely unusual identifier stays confined to
// its own segment instead of adding path levels or spilling into the query
// string (issue #6).
func pathSeg(s string) string {
	return url.PathEscape(s)
}

// pathSegs escapes a slash-bearing path argument (repository file paths)
// segment by segment: the separating slashes are preserved — the server
// must keep seeing distinct segments — while every other reserved character
// is percent-escaped. The result is an escapedPath so pathf inserts it
// verbatim.
func pathSegs(s string) escapedPath {
	parts := strings.Split(s, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return escapedPath(strings.Join(parts, "/"))
}

// pathf builds an endpoint path from a fmt-style format string, escaping
// every string argument as a single path segment. It is the single place
// where endpoint paths are assembled (issue #6: 383 call sites used to
// concatenate caller-controlled values raw).
//
// Rules:
//   - Verbs before the first "?" in the format belong to the path: string
//     args are escaped with url.PathEscape (pathSeg); non-string args
//     (ints via %d) are formatted as-is.
//   - Verbs after the first "?" belong to the query string and are left
//     untouched — callers pass pre-encoded query values (toQuery, Encode).
//   - Arguments of type escapedPath (from pathSegs) are inserted verbatim,
//     never double-escaped.
func pathf(format string, args ...any) string {
	if qi := strings.IndexByte(format, '?'); qi >= 0 {
		escaped := slices.Clone(args)
		escapePathVerbs(format[:qi], escaped)
		return fmt.Sprintf(format, escaped...)
	}
	escapePathVerbs(format, args)
	return fmt.Sprintf(format, args...)
}

// escapePathVerbs escapes, in place, the string arguments bound to the
// format verbs contained in the given (path-only) portion of a format
// string.
func escapePathVerbs(pathFormat string, args []any) {
	arg := 0
	for i := 0; i < len(pathFormat); i++ {
		if pathFormat[i] != '%' || i+1 >= len(pathFormat) {
			continue
		}
		switch pathFormat[i+1] {
		case '%':
			i++
		case 's', 'd', 'v', 'q', 'x', 'X', 'o', 'b', 'f', 'g', 'e':
			if arg < len(args) {
				switch v := args[arg].(type) {
				case escapedPath:
					// already escaped (pathSegs); keep verbatim
				case string:
					args[arg] = pathSeg(v)
				}
			}
			arg++
			i++
		}
	}
}
