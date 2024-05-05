package ethclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Quantity_Encoding(t *testing.T) {
	type V struct {
		V Quantity `json:"V"`
	}

	tests := []struct {
		s string
		v V
	}{
		{s: `{"V":"0x0"}`, v: V{V: NewQuantityFromInt64(0)}},
		{s: `{"V":"0x41"}`, v: V{V: NewQuantityFromInt64(65)}},
		{s: `{"V":"0x400"}`, v: V{V: NewQuantityFromInt64(1024)}},
	}
	for _, tc := range tests {
		t.Run(tc.s+"_encode", func(t *testing.T) {
			v, err := json.Marshal(tc.v)
			assert.NoError(t, err)
			assert.Equal(t, tc.s, string(v))
		})

		t.Run(tc.s+"_decode", func(t *testing.T) {
			var v V
			err := json.Unmarshal([]byte(tc.s), &v)
			assert.NoError(t, err)
			assert.Equal(t, tc.v, v)
		})
	}
}
