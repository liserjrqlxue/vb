package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// warp of *git.Repository
type Repo struct {
	*git.Repository
	TagsMap map[plumbing.Hash]*plumbing.Reference
}

func main() {
	var cwd = HandleError(os.Getwd()).(string)
	var r = Repo{HandleError(git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{DetectDotGit: true})).(*git.Repository), make(map[plumbing.Hash]*plumbing.Reference)}
	var branch = r.branchShowCurrent()
	var tag = r.Describe()
	fmt.Printf("%s:%s\n", branch, tag)

}

func (r *Repo) head() *plumbing.Reference {
	return HandleError(r.Head()).(*plumbing.Reference)
}

func (r *Repo) branchShowCurrent() string {
	return strings.TrimPrefix(r.head().Name().String(), "refs/heads/")
}

func (r *Repo) getTagMap() {
	CheckErr(
		HandleError(
			r.Tags(),
		).(storer.ReferenceIter).
			ForEach(
				func(t *plumbing.Reference) error {
					r.TagsMap[t.Hash()] = t
					return nil
				},
			),
	)
}

// Describe the reference as 'git describe --tags' will do
func (r *Repo) Describe() string {
	// Build the tag map
	r.getTagMap()

	var tag *plumbing.Reference
	var count int
	CheckErr(
		// Fetch the reference log
		HandleError(
			r.Log(
				&git.LogOptions{
					From:  r.head().Hash(),
					Order: git.LogOrderCommitterTime,
				},
			),
		).(object.CommitIter).
			// Search the tag
			ForEach(
				func(c *object.Commit) error {
					if t, ok := r.TagsMap[c.Hash]; ok {
						tag = t
					}
					if tag != nil {
						return storer.ErrStop
					}
					count++
					return nil
				},
			),
	)
	if count == 0 {
		return fmt.Sprint(tag.Name().Short())
	}
	if tag == nil {
		return fmt.Sprintf(
			"%v-%v-g%v",
			"v0.0.0",
			count,
			r.head().Hash().String()[0:7],
		)
	}
	return fmt.Sprintf(
		"%v-%v-g%v",
		tag.Name().Short(),
		count,
		r.head().Hash().String()[0:7],
	)
}

// CheckErr log.Fatal if error
func CheckErr(err error, msg ...string) {
	if err != nil {
		//panic(err)
		log.Fatal(err, msg)
	}
}

// HandleError CheckErr and return as interface
func HandleError(a interface{}, err error) interface{} {
	CheckErr(err)
	return a
}
