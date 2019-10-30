package main

import (
	"fmt"
	"strings"
)

const sep = "."

type replacement struct {
	dst, prefixOld, prefixNew string
}

type rewriter map[string]string
type rewriterGroup map[string]rewriter

func newRewriter(cfg []rewriteConfig) rewriter {
	r := make(rewriter)
	for _, v := range cfg {
		r[v.From] = v.To
	}
	return r
}

// restore transforms back value being rewritten before
func (rep replacement) restore(value string) string {
	return rep.prefixOld + strings.TrimPrefix(value, rep.prefixNew)
}

// commonPrefix groups given rewriter by common prefix containing of each but the last node
func (r rewriter) commonPrefix() rewriterGroup {
	qty := len(r)
	result := make(rewriterGroup, qty)

	for rewriteFrom, rewriteTo := range r {
		prefix := r.extractPrefix(strings.TrimSuffix(rewriteFrom, sep))
		if prefix == "" { // from-patterns which consist of one node exactly are not grouped
			continue
		}

		prefix += sep + "*"
		if _, exists := result[prefix]; !exists {
			result[prefix] = make(rewriter, qty)
		}

		result[prefix][rewriteFrom] = rewriteTo
	}

	result[""] = r
	return result
}

func (r rewriter) extractPrefix(value string) string {
	valueParts := strings.Split(value, sep)
	return strings.Join(valueParts[:len(valueParts)-1], sep)
}

func (r rewriter) do(value string) replacement {
	result := replacement{
		dst: value,
	}

	for k, v := range r {
		if strings.HasPrefix(value, k) {
			result.prefixNew = v
			result.prefixOld = k
			result.dst = v + strings.TrimPrefix(value, k)
			break
		}
	}

	return result
}

// maybeSpawnPaths determines whether pathPattern matches any rewriter group (e.g. if path = a.*.b.c.d and rewriter group is a.*.b)
// if it does then function unfolds pathPattern by applying all matching rewriters
func (rg rewriterGroup) maybeSpawnPaths(pathPattern string) []replacement {
	var (
		found      bool
		group      rewriter
		pathPrefix string
	)

	for k, v := range rg {
		if k != "" && strings.HasPrefix(pathPattern, k) {
			found = true
			pathPrefix = k
			group = v
			break
		}
	}

	if found {
		// match has been found - unfolding
		pathParts := strings.Split(pathPattern, sep)
		pathSuffix := strings.Join(pathParts[strings.Count(pathPrefix, sep)+1:], sep)
		result := make([]replacement, 0, len(group))

		for rewriteFrom, rewriteTo := range group {
			dst := pathSuffix
			rep := replacement{
				dst:       dst,
				prefixOld: rewriteFrom,
				prefixNew: rewriteTo,
			}
			if rewriteTo != "" {
				rep.dst = fmt.Sprintf("%s.%s", strings.TrimSuffix(rep.prefixNew, sep), rep.dst)
			}

			result = append(result, rep)
		}

		return result
	} else {
		// match has not been found - leaving data intact
		return []replacement{rg[""].do(pathPattern)}
	}
}
