// Package requests contains API request payload models.
package requests

type ExplainWord struct {
	Language  string
	CleanWord string
	Phonetic  string
	IsReverse bool
}
