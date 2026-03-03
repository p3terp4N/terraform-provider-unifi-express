package device

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
)

var _ base.ResourceModel = &deviceModel{}

type deviceModel struct {
	base.Model
	MAC            types.String `tfsdk:"mac"`
	Name           types.String `tfsdk:"name"`
	Disabled       types.Bool   `tfsdk:"disabled"`
	AllowAdoption  types.Bool   `tfsdk:"allow_adoption"`
	ForgetOnDestoy types.Bool   `tfsdk:"forget_on_destroy"`
	PortOverrides  types.Set    `tfsdk:"port_override"`
}

type portOverrideModel struct {
	Number            types.Int64  `tfsdk:"number"`
	Name              types.String `tfsdk:"name"`
	PortProfileID     types.String `tfsdk:"port_profile_id"`
	OpMode            types.String `tfsdk:"op_mode"`
	PoeMode           types.String `tfsdk:"poe_mode"`
	AggregateNumPorts types.Int64  `tfsdk:"aggregate_num_ports"`
}

// portOverrideAttrTypes returns the attribute type map for the port_override
// nested set, used when constructing types.Set values.
func portOverrideAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"number":              types.Int64Type,
		"name":                types.StringType,
		"port_profile_id":     types.StringType,
		"op_mode":             types.StringType,
		"poe_mode":            types.StringType,
		"aggregate_num_ports": types.Int64Type,
	}
}

func (m *deviceModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var overrides []portOverrideModel
	if !m.PortOverrides.IsNull() && !m.PortOverrides.IsUnknown() {
		diags.Append(m.PortOverrides.ElementsAs(ctx, &overrides, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Deduplicate by port number (same as V1)
	overrideMap := map[int]unifi.DevicePortOverrides{}
	for _, o := range overrides {
		idx := int(o.Number.ValueInt64())
		overrideMap[idx] = unifi.DevicePortOverrides{
			PortIDX:           idx,
			Name:              o.Name.ValueString(),
			PortProfileID:     o.PortProfileID.ValueString(),
			OpMode:            o.OpMode.ValueString(),
			PoeMode:           o.PoeMode.ValueString(),
			AggregateNumPorts: int(o.AggregateNumPorts.ValueInt64()),
		}
	}

	pos := make([]unifi.DevicePortOverrides, 0, len(overrideMap))
	for _, po := range overrideMap {
		pos = append(pos, po)
	}

	return &unifi.Device{
		ID:            m.ID.ValueString(),
		MAC:           m.MAC.ValueString(),
		Name:          m.Name.ValueString(),
		PortOverrides: pos,
	}, diags
}

func (m *deviceModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.Device)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Device, got: %T", i))
		return diags
	}

	m.ID = types.StringValue(other.ID)
	m.MAC = types.StringValue(other.MAC)
	m.Name = types.StringValue(other.Name)
	m.Disabled = types.BoolValue(other.Disabled)

	// Convert port overrides from API response to types.Set
	poModels := make([]portOverrideModel, 0, len(other.PortOverrides))
	for _, po := range other.PortOverrides {
		poModels = append(poModels, portOverrideModel{
			Number:            types.Int64Value(int64(po.PortIDX)),
			Name:              types.StringValue(po.Name),
			PortProfileID:     types.StringValue(po.PortProfileID),
			OpMode:            types.StringValue(po.OpMode),
			PoeMode:           types.StringValue(po.PoeMode),
			AggregateNumPorts: types.Int64Value(int64(po.AggregateNumPorts)),
		})
	}

	if len(poModels) > 0 {
		setValue, setDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: portOverrideAttrTypes()}, poModels)
		diags.Append(setDiags...)
		if !diags.HasError() {
			m.PortOverrides = setValue
		}
	} else {
		m.PortOverrides = types.SetNull(types.ObjectType{AttrTypes: portOverrideAttrTypes()})
	}

	return diags
}
