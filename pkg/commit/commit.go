// Package commit allows parsing commit messages that follow the conventional
// commits format.
package commit

type Commit struct {
	Type            string
	Scope           string
	Subject         string
	Body            string
	Footers         []*Footer
	References      []*Reference
	BreakingChanges []string
	IsBreaking      bool
}

type Footer struct {
	Name  string
	Value string
}

type Reference struct {
	Name  string
	Value string
}
