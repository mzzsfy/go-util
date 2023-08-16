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

func traverseTree(node treeNode, level int) {
    if node.key != "" {
        fmt.Printf("level:%d,Key: %s, Value: %v\n", level, node.key, node.value)
        for _, child := range node.child {
            traverseTree(child, level+1)
        }
    }
}

//遍历树
func TestTree(t *testing.T) {
    //构建树
    tree := buildTree(4)
    //遍历树
    traverseTree(tree, 1)
    //使用seq遍历树
    FromTreeT(tree, func(node treeNode) Seq[treeNode] { return FromSlice(node.child) }).ForEach(func(node treeNode) {
        fmt.Printf("seq1,Key: %s, Value: %v\n", node.key, node.value)
    })
}
