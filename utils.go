package router

import (
	"errors"
	"regexp"
	"sort"
	"strings"
)

func mergeMultipleSlash(path string) string {
	var duplicateSlash bool
	for i := range path {
		if i != 0 && path[i] == Separator && path[i-1] == Separator {
			duplicateSlash = true
			break
		}
	}
	if !duplicateSlash {
		return path
	}

	var (
		buf  = []byte(path)
		prev int
	)
	for i, b := range buf {
		if i == 0 || b != Separator || b != buf[i-1] {
			if i != prev {
				buf[prev] = buf[i]
			}
			prev += 1
		}
	}
	return string(buf[:prev])
}

func cleanPath(path string) string {
	var (
		begin = 0
		end   = len(path)
	)
	for i := range path {
		if path[i] == Separator {
			begin++
		} else {
			break
		}
	}
	for i := end - 1; i >= 0; i-- {
		if path[i] == Separator {
			end--
		} else {
			break
		}
	}
	return mergeMultipleSlash(path[begin:end])
}

func splitBy(s string, sep byte) (string, string) {
	begin := strings.IndexByte(s, sep)
	if begin < 0 {
		return s, ""
	}
	end := begin
	for l := len(s); end+1 < l && s[end+1] == sep; end++ {
	}
	return s[:begin], s[end+1:]
}

func parseNode(currSec, nextSecs string) (Tree, error) {
	var (
		node Tree
		err  error
	)
	switch currSec[0] {
	case '*':
		node.catch, node.regexp, err = parseNameAndRegexp(currSec[1:])
		if err != nil {
			return node, err
		}
		node.nodeType = _NODE_ANY
		if nextSecs != "" {
			return node, errors.New("catch all syntax must be the last segment")
		}
	case ':':
		node.catch, node.regexp, err = parseNameAndRegexp(currSec[1:])
		if err != nil {
			return node, err
		}
		node.nodeType = _NODE_PARAM
	default:
		node.catch = currSec
		node.nodeType = _NODE_STATIC
	}
	return node, err
}

func extendNodesCap(nodes []*Tree, size int, asCap bool) []*Tree {
	l := len(nodes)
	c := l + size
	if asCap {
		l = c
	}
	newNodes := make([]*Tree, l, c)
	copy(newNodes, nodes)
	return newNodes
}

func extendKVsCap(kvs []KeyValue, size int, asCap bool) []KeyValue {
	l := len(kvs)
	c := l + size
	if asCap {
		l = c
	}
	newKvs := make([]KeyValue, l, c)
	copy(newKvs, kvs)
	return newKvs
}

func parseNameAndRegexp(sec string) (string, *regexp.Regexp, error) {
	index := strings.IndexByte(sec, ':')
	if index < 0 {
		return sec, nil, nil
	}
	name := sec[:index]
	r := sec[index+1:]
	if r == "" {
		return name, nil, nil
	}
	reg, err := regexp.Compile(r)
	return name, reg, err
}

func compareNode(n1, n2 *Tree) int {
	if n1.nodeType < n2.nodeType {
		return -1
	}
	if n1.nodeType > n2.nodeType {
		return 1
	}
	if n1.regexp == nil {
		if n2.regexp == nil {
			return 0
		}
		return 1
	}
	if n2.regexp == nil {
		return -1
	}
	if n1.regexp.String() == n2.regexp.String() {
		return 0
	}
	return -1
}

func firstMatchResult(results []MatchResult) MatchResult {
	if len(results) == 0 {
		return MatchResult{}
	}
	return results[0]
}

type sortByLvResults []MatchResult

func (s sortByLvResults) Len() int {
	return len(s)
}

func (s sortByLvResults) Less(i, j int) bool {
	if s[i].lv != s[j].lv {
		return s[i].lv < s[j].lv
	}
	return s[i].seq < s[j].seq
}

func (s sortByLvResults) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func (s sortByLvResults) Append(kvs []KeyValue, handler interface{}, lv int) sortByLvResults {
	return append(s, MatchResult{
		KeyValues: kvs,
		Handler:   handler,
		lv:        lv,
		seq:       len(s),
	})
}

func (s sortByLvResults) Sort() []MatchResult {
	sort.Sort(s)
	return s
}
