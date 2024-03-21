package juju

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/juju/juju/core/constraints"
)

// -----------------------------------------------------------------------------
//                 ConstraintsType
// -----------------------------------------------------------------------------

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringTypable = ConstraintsType{}

type ConstraintsType struct {
	basetypes.StringType
}

func (t ConstraintsType) Equal(o attr.Type) bool {
	other, ok := o.(ConstraintsType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func GetConstraintsValue(value string) (ConstraintsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	v, err := constraints.Parse(value)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Invalid Constraints Format", err.Error()))
		return ConstraintsUnknown(), diags
	}

	return ConstraintsValue{
		StringValue: basetypes.NewStringValue(value),
		Constraints: v,
	}, diags
}

func (t ConstraintsType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diagnostic diag.Diagnostics
	if in.IsNull() {
		// There is no problem with empty constraints.
		return ConstraintsNull(), diagnostic
	}
	if in.IsUnknown() {
		// There is no problem with unknown constraints.
		return ConstraintsUnknown(), diagnostic
	}

	constVal, err := constraints.Parse(in.String())
	if err != nil {
		diagnostic.AddError(
			"invalid constraints format",
			fmt.Sprintf("failed to parse constraints: %v", err),
		)
		return nil, diagnostic
	}
	// Format to ensure that the constraints are in the correct format
	value := ConstraintsValue{
		StringValue: basetypes.NewStringValue(constVal.String()),
		Constraints: constVal,
	}
	return value, diagnostic
}

func (t ConstraintsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t ConstraintsType) ValueType(ctx context.Context) attr.Value {
	return ConstraintsValue{}
}

func (t ConstraintsType) String() string {
	return "ConstraintsType"
}

func (t ConstraintsType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Constraints Type Validation Error",
			fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}
	if _, err := constraints.Parse(in.String()); err != nil {
		diags.AddAttributeError(
			path,
			"Constraints Type Validation Error",
			fmt.Sprintf("Failed to parse constraints: %v", err),
		)
		return diags
	}

	return diags
}

// -----------------------------------------------------------------------------
//                 ConstraintsValue
// -----------------------------------------------------------------------------

var _ basetypes.StringValuable = ConstraintsValue{}

type ConstraintsValue struct {
	basetypes.StringValue
	Constraints constraints.Value
}

func ConstraintsNull() ConstraintsValue {
	return ConstraintsValue{StringValue: basetypes.NewStringNull()}
}

func ConstraintsUnknown() ConstraintsValue {
	return ConstraintsValue{StringValue: basetypes.NewStringUnknown()}
}

// // From:
// // https://github.com/juju/juju/blob/97ee0aefa11e6ca592ae949d903cef073ac858a4/core/constraints/constraints.go#L850C1-L855C2
// var mbSuffixes = map[string]float64{
// 	"M": 1,
// 	"G": 1024,
// 	"T": 1024 * 1024,
// 	"P": 1024 * 1024 * 1024,
// }

// func getSuffixKeys() []string {
// 	keys := make([]string, len(mbSuffixes))
// 	i := 0
// 	for k := range mbSuffixes {
// 		keys[i] = k
// 		i++
// 	}
// 	return keys
// }

// func convertByteSizeStrToInt(str string) (*uint64, error) {
// 	var value uint64
// 	if str != "" {
// 		mult := 1.0
// 		if m, ok := mbSuffixes[str[len(str)-1:]]; ok {
// 			str = str[:len(str)-1]
// 			mult = m
// 		}
// 		val, err := strconv.ParseFloat(str, 64)
// 		if err != nil || val < 0 {
// 			return nil, errors.Errorf("must be a non-negative float with optional M/G/T/P suffix")
// 		}
// 		val *= mult
// 		value = uint64(math.Ceil(val))
// 	}
// 	return &value, nil
// }

// func convertByteSizeIntToStr(valInt uint64) (*string, error) {
// 	var value uint64

// 	suffix := []string{"M", "G", "T", "P"}

// 	suf := 0
// 	finalVal := valInt
// 	for finalVal > 1024 && value < uint64(len(suffix)) {
// 		finalVal = finalVal / 1024
// 		suf++
// 	}
// 	result := strconv.FormatUint(finalVal, 10) + suffix[suf]
// 	return &result, nil
// }

func (t ConstraintsValue) Equal(o attr.Value) bool {
	other, ok := o.(ConstraintsValue)

	if !ok {
		if _, isStr := o.(basetypes.StringValue); !isStr {
			return false
		}
		// We do not have a constraints type but we have a string, check if we have a valid
		// constraints object from Juju
		valOther, err := constraints.Parse(other.String())
		if err != nil {
			return false
		}
		return reflect.DeepEqual(t.Constraints, valOther)
	}

	return reflect.DeepEqual(t.Constraints, other.Constraints)
}

func (t ConstraintsValue) String() string {
	return t.Constraints.String()
}

func (v ConstraintsValue) Type(ctx context.Context) attr.Type {
	return ConstraintsType{}
}
