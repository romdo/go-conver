package commit

import (
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
	}{
		{
			name: "non-convention commit",
			args: args{header: []byte("add user sorting option")},
			want: &Commit{Subject: "add user sorting option"},
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
			name: "multi-line header (LF)",
			args: args{
				header: []byte("feat(user)!: add usersorting\noption"),
			},
			errStr: "header cannot span multiple lines",
		},
		{
			name: "multi-line header (CR)",
			args: args{
				header: []byte("feat(user)!: add usersorting\roption"),
			},
			errStr: "header cannot span multiple lines",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeader(tt.args.header)

			if tt.errStr != "" {
				assert.Error(t, err, tt.errStr)
			} else {
				assert.Equal(t, tt.want, got)
			}
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
			name: "multiple paragraphs",
			args: args{input: []byte(
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n" +
					"Praesent eleifend lorem non purus finibus, interdum\n" +
					"hendrerit sem bibendum.\n" +
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
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n" +
						"Praesent eleifend lorem non purus finibus, interdum\n" +
						"hendrerit sem bibendum.",
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
			name: "CRLF line separator",
			args: args{input: []byte(
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit.\r\n" +
					"Praesent eleifend lorem non purus finibus, interdum\r\n" +
					"hendrerit sem bibendum.\r\n" +
					"\r\n" +
					"Etiam porttitor mollis nulla, egestas facilisis nisi\r\n" +
					"molestie ut. Quisque mi mi, commodo ut mattis a,\r\n" +
					"scelerisque eu elit.\r\n",
			)},
			want: [][]byte{
				[]byte(
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n" +
						"Praesent eleifend lorem non purus finibus, interdum\n" +
						"hendrerit sem bibendum.",
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
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit.\r" +
					"Praesent eleifend lorem non purus finibus, interdum\r" +
					"hendrerit sem bibendum.\r" +
					"\r" +
					"Etiam porttitor mollis nulla, egestas facilisis nisi\r" +
					"molestie ut. Quisque mi mi, commodo ut mattis a,\r" +
					"scelerisque eu elit.\r",
			)},
			want: [][]byte{
				[]byte(
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n" +
						"Praesent eleifend lorem non purus finibus, interdum\n" +
						"hendrerit sem bibendum.",
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
