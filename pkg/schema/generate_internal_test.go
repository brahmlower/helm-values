package schema

import (
	"testing"

	"helmvalues/pkg"

	"github.com/stretchr/testify/assert"
)

func TestMarkGenerated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing string
		want     string
	}{
		{
			name:     "no comment",
			existing: "",
			want:     GeneratedComment,
		},
		{
			name:     "author comment is kept after the disclaimer",
			existing: "maintainer note",
			want:     GeneratedComment + "\n\nmaintainer note",
		},
		{
			name:     "already marked is unchanged",
			existing: GeneratedComment + "\n\nmaintainer note",
			want:     GeneratedComment + "\n\nmaintainer note",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := pkg.NewJsonSchema()
			s.Comment = tc.existing

			markGenerated(s)

			assert.Equal(t, tc.want, s.Comment)
		})
	}
}
