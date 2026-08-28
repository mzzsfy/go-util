package seq

import (
	"fmt"
	"math/rand"
	"testing"
)

type treeNode struct {
	key   string
	value any
	child []treeNode
}

func buildTree(levels int) treeNode {
	if levels == 1 {
		return treeNode{}
	}

	node := treeNode{key: "root", value: rand.Intn(100)}
	for i := 0; i < 1+rand.Intn(10); i++ {
		key := fmt.Sprintf("node%d", i)
		child := buildTree(levels - 1)
		if child.key != "" {
			node.child = append(node.child, child)
		} else {
			node.child = append(node.child, treeNode{key: key, value: rand.Intn(100)})
		}
	}
	return node
}

// 遍历树,seq遍历与手动遍历结构一致性
func TestTree(t *testing.T) {
	//构建树
	tree := buildTree(4)
	//手动递归计数作为基准
	nodes := 0
	var countNodes func(node treeNode)
	countNodes = func(node treeNode) {
		if node.key != "" {
			nodes++
			for _, child := range node.child {
				countNodes(child)
			}
		}
	}
	countNodes(tree)
	//seq遍历计数,校验首节点为root
	seqNodes := 0
	firstKey := ""
	FromTreeT(tree, func(node treeNode) Seq[treeNode] { return FromSlice(node.child) }).ForEach(func(node treeNode) {
		if seqNodes == 0 {
			firstKey = node.key
		}
		seqNodes++
	})
	if seqNodes != nodes {
		t.Fatalf("遍历节点数不一致:seq=%d manual=%d", seqNodes, nodes)
	}
	if firstKey != "root" {
		t.Fatalf("首节点应为root,实际%s", firstKey)
	}
}
