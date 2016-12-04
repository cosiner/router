package router

import "testing"

func TestTreeNode(t *testing.T) {
	var tree Tree

	t.Log(tree.Add("/*allpath:.*\\.txt", "CatchAll:Named"))
	t.Log(tree.Add("/*allpath:.*\\.txt", func(old interface{}) (interface{}, error) {
		return old.(string) + "Double", nil
	}))
	t.Log(tree.Add("/*path", "CatchAll"))
	t.Log(tree.Add("/:parent/child", "Param:Named"))
	t.Log(tree.Add("/:/:child", "Param"))
	t.Log(tree.Add("/::a/:child", "Param"))
	//tree.Add("/*:.*\\.md", "CatchAllRegexp")
	//tree.Add("/*allpath:.*\\.txt", "CatchAll:NamedRegexp")

	t.Log(tree.MatchAll("/a.txt"))
	t.Log(tree.MatchAll("/a.md"))
	t.Log(tree.MatchAll("/a/childa"))
}
