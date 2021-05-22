package boolstr

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// BoolOrString is a type that can hold an bool or a string.

// +protobuf=true
// +protobuf.options.(gogoproto.goproto_stringer)=false
// +k8s:openapi-gen=true
type BoolOrString struct {
	Type    Type   `protobuf:"varint,1,opt,name=type,casttype=Type"`
	BoolVal bool   `protobuf:"bytes,2,opt,name=boolVal"`
	StrVal  string `protobuf:"bytes,3,opt,name=strVal"`
}

// Type represents the stored type of BoolOrString
type Type int64

const (
	Bool   Type = iota // The BoolOrString holds a bool.
	String             // The BoolOrString holds a string.
)

// FromBool creates a BoolOrString object with a bool value.
func FromBool(val bool) BoolOrString {
	return BoolOrString{Type: Bool, BoolVal: val}
}

// FromString creates a BoolOrString object with a string value.
func FromString(val string) BoolOrString {
	return BoolOrString{Type: String, StrVal: val}
}

// Parse the given string and try to convert it to a bool before
// setting it as a string value.
func Parse(val string) BoolOrString {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return FromString(val)
	}
	return FromBool(b)
}

// UnmarshalJSON implements the json.Unmarshaller interface.
// Booleans will be marshalled to a BoolOrString object of
// Type Bool, as will strings equal to "true" or "false".
// All other values will attempt to be unmarshalled to a
// BoolOrString object of Type String.
func (boolstr *BoolOrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		var strVal string
		if err := json.Unmarshal(value, &strVal); err != nil {
			return err
		}
		*boolstr = Parse(strVal)
		return nil
	}
	boolstr.Type = Bool
	return json.Unmarshal(value, &boolstr.BoolVal)
}

// MarshalJSON implements the json.Marshaller interface.
func (boolstr BoolOrString) MarshalJSON() ([]byte, error) {
	switch boolstr.Type {
	case Bool:
		return json.Marshal(boolstr.BoolVal)
	case String:
		return json.Marshal(boolstr.StrVal)
	default:
		return []byte{}, fmt.Errorf("impossible BoolOrString.Type")
	}
}

// String returns the StrVal if Type String, or the FormatBool of
// the BoolVal if Type Bool. If *BoolOrString is nil, a string
// representation of nil is returned.
func (boolstr *BoolOrString) String() string {
	if boolstr == nil {
		return "<nil>"
	}
	if boolstr.Type == String {
		return boolstr.StrVal
	}
	return strconv.FormatBool(boolstr.BoolVal)
}

// BoolValue returns the BoolVal if Type Bool, or of it is a String,
// will attempt a conversion to bool, returning false if a parsing
// error occurs.
// If *BoolOrString is nil, false is returned.
func (boolstr *BoolOrString) BoolValue() bool {
	if boolstr == nil {
		return false
	}
	if boolstr.Type == String {
		b, _ := strconv.ParseBool(boolstr.StrVal)
		return b
	}
	return boolstr.BoolVal
}
