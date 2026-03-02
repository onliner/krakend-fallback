package main

import "strings"

func MatchPathTemplate(tpl, path string) bool {
	tSeg := splitPath(tpl)
	pSeg := splitPath(path)

	if len(tSeg) != len(pSeg) {
		return false
	}

	for i := range tSeg {
		ts := tSeg[i]
		ps := pSeg[i]

		if isParam(ts) {
			if ps == "" {
				return false
			}
			continue
		}

		if ts != ps {
			return false
		}
	}

	return true
}

func splitPath(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "/")
}

func isParam(seg string) bool {
	return len(seg) >= 3 && strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}
