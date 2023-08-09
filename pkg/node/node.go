package node

import (
	"errors"
	"fmt"
	"os"
)

type Tree struct {
	Root         *Node
	TimeAccessed int64
}

type Node struct {
	Path     string
	IsDir    bool
	Children []*Node
}

func GetPathToNodeMappings(nodes []*Node) map[string]*Node {
	mappings := map[string]*Node{}

	for _, node := range nodes {
		mappings[node.Path] = node
	}

	return mappings
}

type DNode struct {
	Path  string
	Dirs  []*DNode
	Files []*FNode
}

func NewDNode(Path string) (*DNode, error) {
	if Path == "" {
		return nil, errors.New("error creating DNode")
	}

	dnode := &DNode{
		Path:  Path,
		Dirs:  []*DNode{},
		Files: []*FNode{},
	}

	return dnode, nil
}

type Options struct {
	Path   string
	Size   int
	Chunks []string
}

type FNode struct {
	Path   string
	Size   int
	Chunks []string
}

func NewFNode(opt Options) (*FNode, error) {
	Path, Size, Chunks := opt.Path, opt.Size, opt.Chunks

	if Path == "" || Size == 0 || Chunks == nil {
		return nil, errors.New("error creating FNode")
	}

	fnode := &FNode{
		Path:   Path,
		Size:   Size,
		Chunks: Chunks,
	}

	return fnode, nil
}

func walk(node *Node) {
	files, _ := os.ReadDir(node.Path)

	for _, file := range files {
		newNode := &Node{
			Path:     node.Path + "/" + file.Name(),
			IsDir:    false,
			Children: []*Node{},
		}

		if file.IsDir() {
			newNode.IsDir = true
			walk(newNode)
		}

		node.Children = append(node.Children, newNode)
	}
}

func LoadHierarchy(root string) (*Node, error) {
	fileInfo, err := os.Stat(root)

	if err != nil {
		return nil, err
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	rootNode := &Node{
		Path:     root,
		IsDir:    true,
		Children: []*Node{},
	}

	walk(rootNode)

	return rootNode, nil
}
