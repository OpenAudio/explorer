package indexers

import v1 "github.com/OpenAudio/go-openaudio/pkg/api/core/v1"

type Indexer interface {
	GetStatus() *Status
	IndexBatch(blocks *v1.GetBlocksResponse) error
}

type Status struct {
	Name    string `json:"name"`
	Head    uint64 `json:"head"`
	Indexed uint64 `json:"indexed"`
}

func (s *Status) BlockDiff() uint64 {
	return s.Head - s.Indexed
}
