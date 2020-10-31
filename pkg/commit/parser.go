package commit

import (
	"bytes"
	"errors"
	"regexp"
)

const (
	cr   = 13
	lf   = 10
	crlf = "\r\n"
)

var (
	rHeader = regexp.MustCompile(
		`^([\w\-]*)(?:\(([\w\$\.\/\-\* ]*)\))?(!)?\: (.*)$`,
	)
	rFooterToken = regexp.MustCompile(
		`^([\w-]+|BREAKING CHANGE):\s\s*(.*)$`,
	)
	rFooterTicket = regexp.MustCompile(
		`^([\w-]+)\s+(#\S.*)$`,
	)
)

func parseHeader(header []byte) (*Commit, error) {
	if bytes.ContainsAny(header, crlf) {
		return nil, errors.New("header cannot span multiple lines")
	}

	result := rHeader.FindSubmatch(header)
	if result == nil {
		return &Commit{Subject: string(header)}, nil
	}

	return &Commit{
		Type:       string(result[1]),
		Scope:      string(result[2]),
		Subject:    string(result[4]),
		IsBreaking: (string(result[3]) == "!"),
	}, nil
}

func footers(paragraph []byte) []*Footer {
	footers := []*Footer{}
	lines := bytes.Split(paragraph, []byte{lf})

	if !rFooterToken.Match(lines[0]) && !rFooterTicket.Match(lines[0]) {
		return footers
	}

	var cName string
	var cBody []byte
	var cRef bool
	for _, line := range lines {
		if m := rFooterToken.FindSubmatch(line); m != nil {
			if cName != "" {
				footers = append(footers, &Footer{
					Name:      cName,
					Body:      string(bytes.TrimSpace(cBody)),
					Reference: cRef,
				})
				cName = ""
				cRef = false
			}
			cName = string(m[1])
			cBody = m[2]
			cRef = false
		} else if m := rFooterTicket.FindSubmatch(line); m != nil {
			if cName != "" {
				footers = append(footers, &Footer{
					Name:      cName,
					Body:      string(bytes.TrimSpace(cBody)),
					Reference: cRef,
				})
				cName = ""
				cRef = false
			}
			cName = string(m[1])
			cBody = m[2]
			cRef = true
		} else {
			cBody = append(cBody, []byte{lf}...)
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
