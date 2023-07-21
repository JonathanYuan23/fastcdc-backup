package fnode

type Options struct {
	Path   string
	Size   int
	Chunks []string
}

type FNodeError struct {}

func (f *FNodeError) Error() string {
	return "Error creating FNode"
}

type FNode struct {
	Path   string
	Size   int
	Chunks []string
}

func NewFNode(opt Options) (*FNode, error) {
	Path, Size, Chunks := opt.Path, opt.Size, opt.Chunks

	if Path == "" || Size == 0 || Chunks == nil {
		return nil, &FNodeError{}
	}

	fnode := &FNode {
		Path: Path,
		Size: Size,
		Chunks: Chunks,
	}

	return fnode, nil
}
