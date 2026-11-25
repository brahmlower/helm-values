package pkg

import (
	"fmt"
	"math"

	"go.yaml.in/yaml/v4"
)

// Number is a JSON Schema number keyword (minimum, maximum, multipleOf, ...).
// It accepts any YAML int or float and rejects NaN and infinities, which
// JSON cannot represent.
type Number float64

// UnmarshalYAML decodes node as a float64 and rejects non-finite values.
// yaml attaches the node's line to the returned error.
func (n *Number) UnmarshalYAML(node *yaml.Node) error {
	var f float64

	err := node.Decode(&f)
	if err != nil {
		return err //nolint:wrapcheck // keep yaml's LoadErrors unwrapped so line marks survive
	}

	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("%q is not a finite number", node.Value)
	}

	*n = Number(f)

	return nil
}
