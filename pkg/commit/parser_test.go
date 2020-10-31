package commit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
