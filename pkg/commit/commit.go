// Package commit allows parsing commit messages that follow the conventional
// commits format.
package commit

type Commit struct {
	Type       string
	Scope      string
	Subject    string
	Body       string
	Footers    []*Footer
	IsBreaking bool
}

type Footer struct {
	Name string
	Body string
}
