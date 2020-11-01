package commit

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
)

const (
	cr   = 13
	lf   = 10
	crlf = "\r\n"

	typeMatch  = `^[\w-]+$`
	scopeMatch = `^[\w\$\.\/\-\* ]+$`
)

var (
	rHeader = regexp.MustCompile(
		`^([^\(\)]*?)(\((.*?)\))?(!)?\:\s+(.*)$`,
	)
	rFooter = regexp.MustCompile(
		`^([\w-]+)\s+(#.*)|([\w-]+|BREAKING CHANGE):\s+(.*)$`,
	)
	rType  = regexp.MustCompile(typeMatch)
	rScope = regexp.MustCompile(scopeMatch)

	Err = errors.New("")

	ErrFormat          = fmt.Errorf("%winvalid format", Err)
	ErrMultiLineHeader = fmt.Errorf("%w: header has multiple lines", ErrFormat)

	ErrType        = fmt.Errorf("%wtype", Err)
	ErrTypeFormat  = fmt.Errorf("%w must match: %s", ErrType, typeMatch)
	ErrTypeMissing = fmt.Errorf("%w is missing", ErrType)

	ErrScope       = fmt.Errorf("%wscope", Err)
	ErrScopeFormat = fmt.Errorf("%w must match: %s", ErrScope, scopeMatch)
)

func parseHeader(header []byte) (*Commit, error) {
	commit := &Commit{}

	if bytes.ContainsAny(header, crlf) {
		return commit, ErrMultiLineHeader
	}

	result := rHeader.FindSubmatch(header)
	if result == nil {
		commit = &Commit{Subject: string(header)}
	} else {
		commit = &Commit{
			Type:       string(bytes.TrimSpace(result[1])),
			Scope:      string(bytes.TrimSpace(result[3])),
			Subject:    string(bytes.TrimSpace(result[5])),
			IsBreaking: string(result[4]) == "!",
		}
	}

	if commit.Type == "" {
		return commit, ErrTypeMissing
	} else if !rType.MatchString(commit.Type) {
		return commit, ErrTypeFormat
	}

	if len(commit.Scope) > 0 && !rScope.MatchString(commit.Scope) {
		return commit, ErrScopeFormat
	}

	return commit, nil
}

type rawFooter struct {
	Name []byte
	Body []byte
	Ref  bool
}

func footers(paragraph []byte) []*rawFooter {
	footers := []*rawFooter{}
	lines := bytes.Split(bytes.TrimSpace(paragraph), []byte{lf})

	if !rFooter.Match(lines[0]) {
		return footers
	}

	footer := &rawFooter{}
	for _, line := range lines {
		if m := rFooter.FindSubmatch(line); m != nil {
			if len(footer.Name) > 0 {
				footers = append(footers, footer)
			}

			footer = &rawFooter{}
			if len(m[1]) > 0 {
				footer.Name = m[1]
				footer.Body = m[2]
				footer.Ref = true
			} else if len(m[3]) > 0 {
				footer.Name = m[3]
				footer.Body = m[4]
				footer.Ref = false
			}
		} else if len(footer.Name) > 0 {
			footer.Body = append(footer.Body, lf)
			footer.Body = append(footer.Body, line...)
		}
	}

	if len(footer.Name) > 0 {
		footers = append(footers, footer)
	}

	return footers
}

func paragraphs(commitMsg []byte) [][]byte {
	paras := bytes.Split(
		bytes.TrimSpace(normlizeLinefeeds(commitMsg)),
		[]byte{lf, lf},
	)

	for i, p := range paras {
		paras[i] = bytes.TrimSpace(p)
	}

	return paras
}

func normlizeLinefeeds(input []byte) []byte {
	return bytes.ReplaceAll(
		bytes.ReplaceAll(input, []byte{cr, lf}, []byte{lf}),
		[]byte{cr}, []byte{lf},
	)
}
