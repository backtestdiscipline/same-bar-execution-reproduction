package study

import (
	"fmt"
	"io"
	"slices"

	"research-platform/seal"
)

// Inputs contains only source references bound to the run being evaluated or
// verified. OpenSource verifies bytes at EOF; always check its Close error.
type Inputs struct {
	Sources   []seal.ArtifactRef
	workspace *seal.Workspace
}

func NewInputs(w *seal.Workspace, sources []seal.ArtifactRef) Inputs {
	return Inputs{Sources: slices.Clone(sources), workspace: w}
}

func (in Inputs) OpenSource(ref seal.ArtifactRef) (io.ReadCloser, error) {
	if in.workspace == nil || !slices.Contains(in.Sources, ref) {
		return nil, fmt.Errorf("source is not bound to this run")
	}
	return in.workspace.OpenSource(ref)
}
