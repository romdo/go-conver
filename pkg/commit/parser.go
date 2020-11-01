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

func footers(paragraph []byte) []*Footer {
	footers := []*Footer{}
	lines := bytes.Split(paragraph, []byte{lf})

	if !rFooter.Match(lines[0]) {
		return footers
	}

	var cName string
	var cBody []byte
	var cRef bool
	for _, line := range lines {
		if m := rFooter.FindSubmatch(line); m != nil {
			if cName != "" {
				footers = append(footers, &Footer{
					Name:      cName,
					Body:      string(bytes.TrimSpace(cBody)),
					Reference: cRef,
				})
			}
			if len(m[1]) > 0 {
				cName = string(m[1])
				cBody = m[2]
				cRef = true
			} else if len(m[3]) > 0 {
				cName = string(m[3])
				cBody = m[4]
				cRef = false
			}
		} else if cName != "" {
			cBody = append(cBody, lf)
			cBody = append(cBody, line...)
		}
	}

	if cName != "" {
		footers = append(footers, &Footer{
			Name:      cName,
			Body:      string(bytes.TrimSpace(cBody)),
			Reference: cRef,
		})
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
