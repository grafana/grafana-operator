package boolstr

import (
	"encoding/json"
	"testing"

	"github.com/ghodss/yaml"
)

func TestFromBool(t *testing.T) {
	b := FromBool(true)
	if b.Type != Bool || b.BoolVal != true {
		t.Errorf("Expected Type=Bool and BoolVal=True, got %+v", b)
	}
}

func TestFromString(t *testing.T) {
	b := FromString("true")
	if b.Type != String || b.StrVal != "true" {
		t.Errorf("Expected Type=String and StrVal=\"true\", got %+v", b)
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		input  string
		result BoolOrString
	}{
		{"true", FromBool(true)},
		{"false", FromBool(false)},
		{"grafana", FromString("grafana")},
	}

	for _, c := range cases {
		result := Parse(c.input)
		if result != c.result {
			t.Errorf("Failed to parse input '%v': expected %+v, got %+v", c.input, c.result, result)
		}
	}
}

type BoolOrStringHolder struct {
	BOrS BoolOrString `json:"val"`
}

func TestIntOrStringUnmarshalJSON(t *testing.T) {
	cases := []struct {
		input  string
		result BoolOrString
	}{
		{"{\"val\":true}", FromBool(true)},
		{"{\"val\":false}", FromBool(false)},
		{"{\"val\":\"true\"}", FromString("true")},
		{"{\"val\":\"false\"}", FromString("false")},
	}

	for _, c := range cases {
		var result BoolOrStringHolder
		if err := json.Unmarshal([]byte(c.input), &result); err != nil {
			t.Errorf("Failed to unmarshal input '%v': %v", c.input, err)
		}
		if result.BOrS != c.result {
			t.Errorf("Failed to unmarshal input '%v': expected %+v, got %+v", c.input, c.result, result.BOrS)
		}
	}
}

func TestIntOrStringMarshalJSON(t *testing.T) {
	cases := []struct {
		input  BoolOrString
		result string
	}{
		{FromBool(true), "{\"val\":true}"},
		{FromBool(false), "{\"val\":false}"},
		{FromString("true"), "{\"val\":\"true\"}"},
		{FromString("false"), "{\"val\":\"false\"}"},
	}

	for _, c := range cases {
		input := BoolOrStringHolder{c.input}
		result, err := json.Marshal(&input)
		if err != nil {
			t.Errorf("Failed to marshal input '%+v': %v", c.input, err)
		}
		if string(result) != c.result {
			t.Errorf("Failed to marshal input '%+v': expected %v, got %v", c.input, c.result, string(result))
		}
	}
}

func TestIntOrStringUnMarshalYAML(t *testing.T) {
	cases := []struct {
		input  string
		result BoolOrString
	}{
		{"val: true\n", FromBool(true)},
		{"val: false\n", FromBool(false)},
		{"val: \"true\"\n", FromString("true")},
		{"val: \"false\"\n", FromString("false")},
	}

	for _, c := range cases {
		var result BoolOrStringHolder
		if err := yaml.Unmarshal([]byte(c.input), &result); err != nil {
			t.Errorf("Failed to unmarshal input '%v': %v", c.input, err)
		}
		if result.BOrS != c.result {
			t.Errorf("Failed to unmarshal input '%v': expected %+v, got %+v", c.input, c.result, result.BOrS)
		}
	}
}

func TestIntOrStringMarshalYAML(t *testing.T) {
	cases := []struct {
		input  BoolOrString
		result string
	}{
		{FromBool(true), "val: true\n"},
		{FromBool(false), "val: false\n"},
		{FromString("true"), "val: \"true\"\n"},
		{FromString("false"), "val: \"false\"\n"},
	}

	for _, c := range cases {
		input := BoolOrStringHolder{c.input}
		result, err := yaml.Marshal(&input)
		if err != nil {
			t.Errorf("Failed to marshal input '%+v': %v", c.input, err)
		}
		if string(result) != c.result {
			t.Errorf("Failed to marshal input '%+v': expected %v, got %v", c.input, c.result, string(result))
		}
	}
}
