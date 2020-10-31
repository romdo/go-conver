// Package commit allows parsing commit messages that follow the conventional
// commits format.
package commit

type Commit struct {
	Type        string
	Scope       string
	Description string
	Body        string
	Footers     []Footer
}

type Footer struct {
	Name string
	Body string
}
