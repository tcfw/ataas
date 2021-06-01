package js

import (
	"bytes"
	"io"

	babel "github.com/jvatic/goja-babel"
)

var (
	DefaultOpts = map[string]interface{}{
		"plugins": []interface{}{
			[]interface{}{"transform-template-literals", map[string]interface{}{"loose": false, "spec": false}},
			"transform-literals",
			"transform-function-name",
			[]interface{}{"transform-arrow-functions", map[string]interface{}{"spec": false}},
			"transform-block-scoped-functions",
			[]interface{}{"transform-classes", map[string]interface{}{"loose": false}},
			"transform-object-super",
			"transform-shorthand-properties",
			"transform-duplicate-keys",
			[]interface{}{"transform-computed-properties", map[string]interface{}{"loose": false}},
			"transform-for-of",
			"transform-sticky-regex",
			"transform-unicode-regex",
			[]interface{}{"transform-spread", map[string]interface{}{"loose": false}},
			"transform-parameters",
			[]interface{}{"transform-destructuring", map[string]interface{}{"loose": false}},
			"transform-block-scoping",
			"transform-typeof-symbol",
			// all the other module plugins are just dropped
			[]interface{}{"transform-modules-commonjs", map[string]interface{}{"loose": false}},
			"transform-regenerator",
			"transform-exponentiation-operator",
			"transform-async-to-generator",
		},
		"ast":           false,
		"sourceMaps":    false,
		"babelrc":       false,
		"compact":       false,
		"retainLines":   true,
		"highlightCode": false,
	}
)

func convertJS(code []byte) ([]byte, error) {
	babel.Init(4)
	res, err := babel.Transform(bytes.NewReader(code), DefaultOpts)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(res)
	if err != nil {
		return nil, err
	}

	return b, nil
}
