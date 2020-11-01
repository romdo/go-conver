package commit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseHeader(t *testing.T) {
	type args struct {
		header []byte
	}
	tests := []struct {
		name   string
		args   args
		want   *Commit
		errStr string
		errIs  []error
	}{
		{
			name:  "missing type",
			args:  args{header: []byte("add user sorting option")},
			want:  &Commit{Subject: "add user sorting option"},
			errIs: []error{Err, ErrTypeMissing},
		},
		{
			name: "missing type with scope",
			args: args{
				header: []byte("(user): add user sorting option"),
			},
			want:   &Commit{Scope: "user", Subject: "add user sorting option"},
			errStr: `type is missing`,
			errIs:  []error{Err, ErrType, ErrTypeMissing},
		},
		{
			name: "type only",
			args: args{header: []byte("feat: add user sorting option")},
			want: &Commit{Type: "feat", Subject: "add user sorting option"},
		},
		{
			name: "type and scope",
			args: args{header: []byte("feat(user): add user sorting option")},
			want: &Commit{
				Type:    "feat",
				Scope:   "user",
				Subject: "add user sorting option",
			},
		},
		{
			name: "type and breaking",
			args: args{header: []byte("feat!: add user sorting option")},
			want: &Commit{
				Type:       "feat",
				Subject:    "add user sorting option",
				IsBreaking: true,
			},
		},
		{
			name: "type, scope and breaking",
			args: args{header: []byte("feat(user)!: add user sorting option")},
			want: &Commit{
				Type:       "feat",
				Scope:      "user",
				Subject:    "add user sorting option",
				IsBreaking: true,
			},
		},
		{
			name: "type with underscore (_)",
			args: args{header: []byte("int_feat: add user sorting option")},
			want: &Commit{
				Type:    "int_feat",
				Subject: "add user sorting option",
			},
		},
		{
			name: "type with hyphen (-)",
			args: args{header: []byte("int-feat: add user sorting option")},
			want: &Commit{
				Type:    "int-feat",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with underscopre (_)",
			args: args{
				header: []byte("feat(user_sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user_sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with hyphen (-)",
			args: args{
				header: []byte("feat(user-sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user-sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with slash (/)",
			args: args{
				header: []byte("feat(user/sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user/sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with period (.)",
			args: args{
				header: []byte("feat(user.sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user.sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with dollar sign ($)",
			args: args{
				header: []byte("feat($user): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "$user",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with star (*)",
			args: args{
				header: []byte("feat(user*): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user*",
				Subject: "add user sorting option",
			},
		},
		{
			name: "scope with space ( )",
			args: args{
				header: []byte("feat(user sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "excess whitespace in type with scope",
			args: args{
				header: []byte("  feat  (user sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "excess whitespace in scope",
			args: args{
				header: []byte("feat(  user sort ): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user sort",
				Subject: "add user sorting option",
			},
		},
		{
			name: "excess whitespace in subject",
			args: args{
				header: []byte("feat(user):   add user sorting option  "),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user",
				Subject: "add user sorting option",
			},
		},
		{
			name: "empty scope",
			args: args{
				header: []byte("feat(): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Subject: "add user sorting option",
			},
		},
		{
			name: "multi-line header (LF)",
			args: args{
				header: []byte("feat(user)!: add user sorting\nnoption"),
			},
			want:   &Commit{},
			errStr: "invalid format: header has multiple lines",
			errIs:  []error{ErrFormat, ErrMultiLineHeader},
		},
		{
			name: "multi-line header (CR)",
			args: args{
				header: []byte("feat(user)!: add user sorting\roption"),
			},
			want:   &Commit{},
			errStr: "invalid format: header has multiple lines",
			errIs:  []error{Err, ErrFormat, ErrMultiLineHeader},
		},
		{
			name: "invalid type character",
			args: args{
				header: []byte("feat/internal: add user sorting option"),
			},
			want: &Commit{
				Type:    "feat/internal",
				Subject: "add user sorting option",
			},
			errStr: `type must match: ^[\w-]+$`,
			errIs:  []error{Err, ErrType, ErrTypeFormat},
		},
		{
			name: "invalid type character with scope",
			args: args{
				header: []byte("feat/internal(user): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat/internal",
				Scope:   "user",
				Subject: "add user sorting option",
			},
			errStr: `type must match: ^[\w-]+$`,
			errIs:  []error{Err, ErrType, ErrTypeFormat},
		},
		{
			name: "invalid scope character",
			args: args{
				header: []byte("feat(user#sort): add user sorting option"),
			},
			want: &Commit{
				Type:    "feat",
				Scope:   "user#sort",
				Subject: "add user sorting option",
			},
			errStr: `scope must match: ^[\w\$\.\/\-\* ]+$`,
			errIs:  []error{Err, ErrScope, ErrScopeFormat},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeader(tt.args.header)

			if tt.errStr != "" {
				assert.EqualError(t, err, tt.errStr)
			}

			if len(tt.errIs) > 0 {
				for _, errIs := range tt.errIs {
					assert.True(t, errors.Is(err, errIs))
				}
			}

			if len(tt.errIs) == 0 && tt.errStr == "" {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_footers(t *testing.T) {
	type args struct {
		paragraph []byte
	}
	tests := []struct {
		name string
		args args
		want []*Footer
	}{
		{
			name: "without footer",
			args: args{[]byte("this is not a fooder")},
			want: []*Footer{},
		},
		{
			name: "token footer on second line",
			args: args{[]byte("this is not a fooder\nDone-By: John")},
			want: []*Footer{},
		},
		{
			name: "ticket footer on second line",
			args: args{[]byte("this is not a fooder\nFixes #42")},
			want: []*Footer{},
		},
		{
			name: "breaking change footer on second line",
			args: args{[]byte("this is not a fooder\nBREAKING CHANGE: Oops")},
			want: []*Footer{},
		},
		{
			name: "token footer",
			args: args{[]byte("Reviewed-By: John Smith")},
			want: []*Footer{{Name: "Reviewed-By", Body: "John Smith"}},
		},
		{
			name: "breaking change footer",
			args: args{[]byte("BREAKING CHANGE: Oopsy")},
			want: []*Footer{{Name: "BREAKING CHANGE", Body: "Oopsy"}},
		},
		{
			name: "ticket footer",
			args: args{[]byte("Fixes #82")},
			want: []*Footer{{Name: "Fixes", Body: "#82", Reference: true}},
		},
		{
			name: "multiple token footers",
			args: args{[]byte(
				"Reviewed-By: John\n" +
					"Committer: Smith\n",
			)},
			want: []*Footer{
				{Name: "Reviewed-By", Body: "John"},
				{Name: "Committer", Body: "Smith"},
			},
		},
		{
			name: "multiple ticket footers",
			args: args{[]byte("Fixes #82\nFixes #74")},
			want: []*Footer{
				{Name: "Fixes", Body: "#82", Reference: true},
				{Name: "Fixes", Body: "#74", Reference: true},
			},
		},
		{
			name: "multiple breaking change footers",
			args: args{[]byte(
				"BREAKING CHANGE: Oopsy\n" +
					"BREAKING CHANGE: Again!",
			)},
			want: []*Footer{
				{Name: "BREAKING CHANGE", Body: "Oopsy"},
				{Name: "BREAKING CHANGE", Body: "Again!"},
			},
		},
		{
			name: "mixture of footer types",
			args: args{[]byte(
				"Fixes #930\n" +
					"BREAKING CHANGE: Careful!\n" +
					"Reviewed-By: Maria\n",
			)},
			want: []*Footer{
				{Name: "Fixes", Body: "#930", Reference: true},
				{Name: "BREAKING CHANGE", Body: "Careful!"},
				{Name: "Reviewed-By", Body: "Maria"},
			},
		},
		{
			name: "multi-line footers",
			args: args{[]byte(
				"Description: Lorem ipsum dolor sit amet, consectetur\n" +
					"adipiscing elit.Praesent eleifend lorem non purus\n" +
					"finibus, interdum hendrerit sem bibendum.\n" +
					"Fixes #94\n" +
					"Misc-Other: Etiam porttitor mollis nulla, egestas\n" +
					"facilisis nisi molestie ut. Quisque mi mi, commodo\n" +
					"ut mattis a, scelerisque eu elit.\n" +
					"BREAKING CHANGE: Duis id nulla eget velit maximus\n" +
					"varius et egestas sem. Ut mi risus, pretium quis\n" +
					"cursus quis, porttitor in ipsum.\n",
			)},
			want: []*Footer{
				{
					Name: "Description",
					Body: "Lorem ipsum dolor sit amet, consectetur\n" +
						"adipiscing elit.Praesent eleifend lorem non purus\n" +
						"finibus, interdum hendrerit sem bibendum.",
				},
				{Name: "Fixes", Body: "#94", Reference: true},
				{
					Name: "Misc-Other",
					Body: "Etiam porttitor mollis nulla, egestas\n" +
						"facilisis nisi molestie ut. Quisque mi mi, commodo\n" +
						"ut mattis a, scelerisque eu elit.",
				},
				{
					Name: "BREAKING CHANGE",
					Body: "Duis id nulla eget velit maximus\n" +
						"varius et egestas sem. Ut mi risus, pretium quis\n" +
						"cursus quis, porttitor in ipsum.",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := footers(tt.args.paragraph)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_paragraph(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "single line string",
			args: args{input: []byte("hello world\n")},
			want: [][]byte{[]byte("hello world")},
		},
		{
			name: "multi-line line string",
			args: args{input: []byte("hello world\nthe brown fox\n")},
			want: [][]byte{[]byte("hello world\nthe brown fox")},
		},
		{
			name: "excess whitespace",
			args: args{input: []byte(" \n hello world\nthe brown fox \n ")},
			want: [][]byte{[]byte("hello world\nthe brown fox")},
		},
		{
			name: "multiple paragraphs",
			args: args{input: []byte(
				"Lorem ipsum dolor sit amet, consectetur adipiscing\n" +
					"elit.Praesent eleifend lorem non purus finibus,\n" +
					"interdum hendrerit sem bibendum.\n" +
					"\n" +
					"Etiam porttitor mollis nulla, egestas facilisis nisi\n" +
					"molestie ut. Quisque mi mi, commodo ut mattis a,\n" +
					"scelerisque eu elit.\n" +
					"\n" +
					"Duis id nulla eget velit maximus varius et egestas\n" +
					"sem. Ut mi risus, pretium quis cursus quis,\n" +
					"porttitor in ipsum.\n",
			)},
			want: [][]byte{
				[]byte(
					"Lorem ipsum dolor sit amet, consectetur adipiscing\n" +
						"elit.Praesent eleifend lorem non purus finibus,\n" +
						"interdum hendrerit sem bibendum.",
				),
				[]byte(
					"Etiam porttitor mollis nulla, egestas facilisis nisi\n" +
						"molestie ut. Quisque mi mi, commodo ut mattis a,\n" +
						"scelerisque eu elit.",
				),
				[]byte(
					"Duis id nulla eget velit maximus varius et egestas\n" +
						"sem. Ut mi risus, pretium quis cursus quis,\n" +
						"porttitor in ipsum.",
				),
			},
		},
		{
			name: "paragraphs with surrounding whitespace",
			args: args{input: []byte(
				"\n" +
					" \n" +
					"   Lorem ipsum dolor sit amet, consectetur adipiscing\n" +
					"elit.Praesent eleifend lorem non purus finibus,\n" +
					"interdum hendrerit sem bibendum.  \n" +
					"\n" +
					"\n" +
					"  Etiam porttitor mollis nulla, egestas facilisis nisi\n" +
					"molestie ut. Quisque mi mi, commodo ut mattis a,\n" +
					"scelerisque eu elit.\n" +
					" \n" +
					" ",
			)},
			want: [][]byte{
				[]byte(
					"Lorem ipsum dolor sit amet, consectetur adipiscing\n" +
						"elit.Praesent eleifend lorem non purus finibus,\n" +
						"interdum hendrerit sem bibendum.",
				),
				[]byte(
					"Etiam porttitor mollis nulla, egestas facilisis nisi\n" +
						"molestie ut. Quisque mi mi, commodo ut mattis a,\n" +
						"scelerisque eu elit.",
				),
			},
		},
		{
			name: "CRLF line separator",
			args: args{input: []byte(
				"Lorem ipsum dolor sit amet, consectetur adipiscing\r\n" +
					"elit.Praesent eleifend lorem non purus finibus,\r\n" +
					"interdum hendrerit sem bibendum.\r\n" +
					"\r\n" +
					"Etiam porttitor mollis nulla, egestas facilisis nisi\r\n" +
					"molestie ut. Quisque mi mi, commodo ut mattis a,\r\n" +
					"scelerisque eu elit.\r\n",
			)},
			want: [][]byte{
				[]byte(
					"Lorem ipsum dolor sit amet, consectetur adipiscing\n" +
						"elit.Praesent eleifend lorem non purus finibus,\n" +
						"interdum hendrerit sem bibendum.",
				),
				[]byte(
					"Etiam porttitor mollis nulla, egestas facilisis nisi\n" +
						"molestie ut. Quisque mi mi, commodo ut mattis a,\n" +
						"scelerisque eu elit.",
				),
			},
		},
		{
			name: "CR line separator",
			args: args{input: []byte(
				"Lorem ipsum dolor sit amet, consectetur adipiscing\r" +
					"elit.Praesent eleifend lorem non purus finibus,\r" +
					"interdum hendrerit sem bibendum.\r" +
					"\r" +
					"Etiam porttitor mollis nulla, egestas facilisis nisi\r" +
					"molestie ut. Quisque mi mi, commodo ut mattis a,\r" +
					"scelerisque eu elit.\r",
			)},
			want: [][]byte{
				[]byte(
					"Lorem ipsum dolor sit amet, consectetur adipiscing\n" +
						"elit.Praesent eleifend lorem non purus finibus,\n" +
						"interdum hendrerit sem bibendum.",
				),
				[]byte(
					"Etiam porttitor mollis nulla, egestas facilisis nisi\n" +
						"molestie ut. Quisque mi mi, commodo ut mattis a,\n" +
						"scelerisque eu elit.",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paragraphs(tt.args.input)

			assert.Equal(t, tt.want, got)
		})
	}
}
