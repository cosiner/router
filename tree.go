package router

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
)

var Separator byte = '/'

type nodeType uint8

const (
	_NODE_ROOT nodeType = iota
	_NODE_STATIC
	_NODE_PARAM
	_NODE_ANY
)

type Tree struct {
	nodeType nodeType
	catch    string
	regexp   *regexp.Regexp
	handler  interface{}

	children []*Tree
}

type KeyValue struct {
	Key   string
	Value string
}

type KeyValues []KeyValue

func (kvs KeyValues) Get(key string) string {
	for i := range kvs {
		if kvs[i].Key == key {
			return kvs[i].Value
		}
	}
	return ""
}

func (kvs KeyValues) Len() int {
	return len(kvs)
}

func (kvs KeyValues) Append(key, value string) KeyValues {
	return append(kvs, KeyValue{Key: key, Value: value})
}

func (kvs KeyValues) ExtendAppend(key, value string) KeyValues {
	return KeyValues(extendKVsCap(kvs, 1, false)).Append(key, value)
}

type MatchResult struct {
	KeyValues KeyValues
	Handler   interface{}
	lv        int
	seq       int
}

func (t *Tree) MatchOne(path string) MatchResult {
	one, _ := t.match(path, true, false)
	return firstMatchResult(one)
}

func (t *Tree) MatchAll(path string) []MatchResult {
	_, all := t.match(path, false, true)
	return all
}

func (t *Tree) MatchBoth(path string) (MatchResult, []MatchResult) {
	one, all := t.match(path, true, true)
	return firstMatchResult(one), all
}

func (t *Tree) match(path string, needOne, needAll bool) ([]MatchResult, []MatchResult) {
	if path == "" {
		return nil, nil
	}

	one, all := t.matchPath(nil, nil, nil, cleanPath(path), needOne, needAll, 0)
	return one, all.Sort()
}

func (t *Tree) matchPath(one, all sortByLvResults, kvs KeyValues, path string, needOne, needAll bool, lv int) (sortByLvResults, sortByLvResults) {
	if t.handler != nil {
		if needOne && path == "" && len(one) == 0 {
			one = one.Append(kvs, t.handler, lv)
		}
		if needAll {
			all = all.Append(kvs, t.handler, lv)
		}
	}

	currSec, nextSecs := splitBy(path, Separator)
	for _, child := range t.children {
		newKvs := kvs
		switch child.nodeType {
		case _NODE_ANY:
			if child.regexp == nil || child.regexp.MatchString(path) {
				if child.catch != "" {
					newKvs = kvs.ExtendAppend(child.catch, path)
					if !needAll {
						kvs = newKvs[:len(newKvs)-1]
					}
				}
				if needOne && len(one) == 0 {
					one = one.Append(newKvs, child.handler, lv)
				}
				if needAll {
					all = all.Append(newKvs, child.handler, lv)
				}
			}
		case _NODE_STATIC:
			if path != "" && child.catch == currSec {
				one, all = child.matchPath(one, all, newKvs, nextSecs, needOne, needAll, lv+1)
			}
		case _NODE_PARAM:
			if path != "" && (child.regexp == nil || child.regexp.MatchString(currSec)) {
				if child.catch != "" {
					newKvs = kvs.ExtendAppend(child.catch, currSec)
					if !needAll {
						kvs = newKvs[:len(newKvs)-1]
					}
				}
				one, all = child.matchPath(one, all, newKvs, nextSecs, needOne, needAll, lv+1)
			}
		}
		if !needAll && len(one) > 0 {
			break
		}
	}

	return one, all
}

func (t *Tree) Add(path string, handler interface{}) error {
	if path == "" || handler == nil {
		return errors.New("illegal path or handler")
	}
	subtree, _ := handler.(*Tree)
	if subtree == nil {
		tr, ok := handler.(Tree)
		if ok {
			subtree = &tr
		}
	}
	fn, _ := handler.(func(interface{}) (interface{}, error))
	return t.addPath(path, handler, subtree, fn)
}

func (t *Tree) Child(path string) (*Tree, error) {
	path = cleanPath(path)
	curr := t
	for path != "" {
		currSec, nextSecs := splitBy(path, Separator)
		node, err := parseNode(currSec, nextSecs)
		if err != nil {
			return nil, err
		}
		curr, err = curr.addChild(&node)
		if err != nil {
			return nil, err
		}
		path = nextSecs
	}
	return curr, nil
}

func (t *Tree) addPath(path string, handler interface{}, subtree *Tree, fn func(interface{}) (interface{}, error)) error {
	child, err := t.Child(path)
	if err != nil {
		return err
	}
	if subtree != nil {
		if subtree.nodeType == _NODE_ROOT {
			for _, c := range subtree.children {
				_, err = child.addChild(c)
				if err != nil {
					break
				}
			}
		} else {
			if child == subtree {
				return fmt.Errorf("sub tree is already added")
			}
			_, err = child.addChild(subtree)
		}
		return err
	}
	if fn != nil {
		child.handler, err = fn(child.handler)
		if err != nil {
			return err
		}
		if child.handler == nil {
			return errors.New("empty handler")
		}
		return nil
	}
	if child.handler != nil {
		return errors.New("duplicate handler")
	}
	child.handler = handler
	return nil
}

func (t *Tree) addChild(node *Tree) (child *Tree, err error) {
	var (
		l      = len(t.children)
		result int
	)
	index := sort.Search(l, func(i int) bool {
		child := t.children[i]
		result = compareNode(child, node)
		return result >= 0
	})
	if index == l {
		t.children = extendNodesCap(t.children, 1, false)
		t.children = append(t.children, node)
	} else {
		child := t.children[index]
		if result == 0 {
			if node.handler != nil && child.handler != nil && child.handler != node.handler {
				return child, errors.New("duplicate path with different handler")
			}
			if child.catch == "" {
				child.catch = node.catch
			}
			if child.catch != "" && node.catch != "" && node.catch != child.catch {
				return child, errors.New("param with different name is not allowed")
			}
		} else {
			t.children = extendNodesCap(t.children, 1, true)
			copy(t.children[index+1:], t.children[index:])
			t.children[index] = node
		}
	}

	child = t.children[index]
	if result == 0 {
		for _, c := range node.children {
			_, err = child.addChild(c)
			if err != nil {
				return child, err
			}
		}
	}
	return child, nil
}

func (t *Tree) printPathTree(ctx string) {
	for _, c := range t.children {
		var catch string
		switch c.nodeType {
		default:
			continue
		case _NODE_STATIC:
			catch = c.catch
		case _NODE_PARAM:
			catch = ":" + c.catch
		case _NODE_ANY:
			catch = "*" + c.catch
		}
		if c.regexp != nil {
			catch += ":" + c.regexp.String()
		}
		curr := ctx + "/" + catch
		if c.handler != nil {
			fmt.Println(curr)
		}
		c.printPathTree(curr)
	}
}

func (t *Tree) PrintPathTree() {
	t.printPathTree("")
}
