package main

import (
	"reflect"
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name          string
		masterConfig  []map[string]interface{}
		appProperties map[string]interface{}
		wantedErr     bool
	}{
		{"happy case",
			[]map[string]interface{}{
				{"config1": "value"},
				{"config2": "some-value"},
			},
			map[string]interface{}{
				"config1": "value",
				"config2": "some-value",
			},
			false,
		},
		{"wrong value",
			[]map[string]interface{}{
				{"config1": "value"},
				{"config2": "some-value"},
			},
			map[string]interface{}{
				"config1": "value",
				"config2": "wrong-value",
			},
			true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := compare(test.masterConfig, test.appProperties)
			if !reflect.DeepEqual(err != nil, test.wantedErr) {
				t.Errorf("compare() got error %v but wanted %v", err, test.wantedErr)
			}
		})
	}
}
