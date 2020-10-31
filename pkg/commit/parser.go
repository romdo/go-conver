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

var rHeader = regexp.MustCompile(
	`^([\w\-]*)(?:\(([\w\$\.\/\-\* ]*)\))?(!)?\: (.*)$`,
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
