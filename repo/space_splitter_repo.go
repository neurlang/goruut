package repo

import (
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type ISpaceSplitterRepository interface {
	Split(string) []string
}
type SpaceSplitterRepository struct {
	repo ISpaceSplitterRepository
}

func (s *SpaceSplitterRepository) Split(sentence string) []string {
	return strings.Fields(sentence)
}

func NewSpaceSplitterRepository(di *DependencyInjection) *SpaceSplitterRepository {

	return &SpaceSplitterRepository{}
}

var _ ISpaceSplitterRepository = &SpaceSplitterRepository{}
