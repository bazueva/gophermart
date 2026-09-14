package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerAddr_Set(t *testing.T) {
	type test struct {
		name string
		addr string
		want ServerAddr
		err  error
	}

	tests := []test{
		{
			name: "empty addr",
			addr: "",
			want: ServerAddr{},
			err:  fmt.Errorf("wrong format"),
		},
		{
			name: "without port and host",
			addr: ":",
			want: ServerAddr{},
			err:  fmt.Errorf("wrong port - strconv.Atoi: parsing \"\": invalid syntax"),
		},
		{
			name: "port not number",
			addr: ":test",
			want: ServerAddr{},
			err:  fmt.Errorf("wrong port - strconv.Atoi: parsing \"test\": invalid syntax"),
		},
		{
			name: "without host",
			addr: ":8080",
			want: ServerAddr{Port: 8080},
		},
		{
			name: "with port and host",
			addr: "localhost:8080",
			want: ServerAddr{Port: 8080, Host: "localhost"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := ServerAddr{}

			err := addr.Set(tt.addr)
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.want, addr)
		})
	}
}
