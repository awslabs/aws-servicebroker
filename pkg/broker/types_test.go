package broker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestCosts(t *testing.T) {
	ydoc := []byte(`
- Amount:
    EUR: 29.91
    USD: 30
  Unit: Monthly
- Amount: 
    EUR: 14
  Unit: Year
`)
	var costs []CfnCost
	err := yaml.Unmarshal(ydoc, &costs)
	assert.Nil(t, err)
	assert.Len(t, costs, 2)

	cost := costs[0]
	assert.EqualValues(t, CfnCost{
		Amount: map[string]float64{
			"EUR": 29.91,
			"USD": 30.0,
		},
		Unit: "Monthly",
	}, cost)
	cost = costs[1]
	assert.EqualValues(t, CfnCost{
		Amount: map[string]float64{
			"EUR": 14.0,
		},
		Unit: "Year",
	}, cost)

	t.Run("JSON Marshall", func(t *testing.T) {
		b, err := json.Marshal(costs)
		assert.Nil(t, err)

		assert.Equal(t, `[{"amount":{"EUR":29.91,"USD":30},"unit":"Monthly"},{"amount":{"EUR":14},"unit":"Year"}]`, string(b))
	})

}
