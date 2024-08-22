package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Replace(t *testing.T) {
	kases := []struct {
		src      string
		expr     string
		repl     string
		expected string
	}{
		{`(/uploads/6d64fdcfa1bc9a8dabce73c9329cf7d2/foo.png)`,
			`\(/uploads/(\w{32}/.*)\)`,
			`(http://foo.com/other-uploads/$1)`,
			`(http://foo.com/other-uploads/6d64fdcfa1bc9a8dabce73c9329cf7d2/foo.png)`},
	}

	for _, kase := range kases {
		out, err := Replace(kase.src, kase.expr, kase.repl)
		assert.Nil(t, err)
		assert.Equal(t, kase.expected, out)
	}
}
