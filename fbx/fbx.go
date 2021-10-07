package fbx

import (
	"fmt"
	"io"
	"os"
)

func Load(path string) (*Document, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Parse(r)
}

func Parse(r io.Reader) (*Document, error) {
	p := binaryParser{r: &positionReader{r: r}}
	root, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return BuildDocument(root)
}

func Save(doc *Document, path string) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	return Write(w, doc)
}

func Write(w io.Writer, doc *Document) error {
	fmt.Fprintln(w, "; FBX 7.5.0 project file")
	fmt.Fprintln(w, "; Generator: https://github.com/binzume/modelconv")
	fmt.Fprintln(w, "; -----------------------------------------------")
	fmt.Fprintln(w, "")
	for _, n := range doc.RawNode.Children {
		if n.Name != "FileId" {
			n.Dump(w, 0, true)
		}
	}
	return nil
}
