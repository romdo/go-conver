package commit

import "bytes"

const (
	cr = 13
	lf = 10
)

func paragraphs(input []byte) [][]byte {
	cln := bytes.ReplaceAll(input, []byte{cr, lf}, []byte{lf})
	cln = bytes.ReplaceAll(cln, []byte{cr}, []byte{lf})

	ps := bytes.Split(cln, []byte{lf, lf})

	for i, p := range ps {
		ps[i] = bytes.Trim(p, "\r\n")
	}

	return ps
}
