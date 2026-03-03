package routing

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &portForwardModel{}

type portForwardModel struct {
	base.Model
	DstPort              types.String `tfsdk:"dst_port"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	FwdIP                types.String `tfsdk:"fwd_ip"`
	FwdPort              types.String `tfsdk:"fwd_port"`
	Log                  types.Bool   `tfsdk:"log"`
	Name                 types.String `tfsdk:"name"`
	PortForwardInterface types.String `tfsdk:"port_forward_interface"`
	Protocol             types.String `tfsdk:"protocol"`
	SrcIP                types.String `tfsdk:"src_ip"`
}

func (m *portForwardModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.PortForward{
		ID:            m.ID.ValueString(),
		DstPort:       m.DstPort.ValueString(),
		Enabled:       m.Enabled.ValueBool(),
		Fwd:           m.FwdIP.ValueString(),
		FwdPort:       m.FwdPort.ValueString(),
		Log:           m.Log.ValueBool(),
		Name:          m.Name.ValueString(),
		PfwdInterface: m.PortForwardInterface.ValueString(),
		Proto:         m.Protocol.ValueString(),
		Src:           m.SrcIP.ValueString(),
	}, diag.Diagnostics{}
}

func (m *portForwardModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.PortForward)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.PortForward")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.DstPort = ut.StringOrNull(other.DstPort)
	m.Enabled = types.BoolValue(other.Enabled)
	m.FwdIP = ut.StringOrNull(other.Fwd)
	m.FwdPort = ut.StringOrNull(other.FwdPort)
	m.Log = types.BoolValue(other.Log)
	m.Name = ut.StringOrNull(other.Name)
	m.PortForwardInterface = ut.StringOrNull(other.PfwdInterface)
	m.Protocol = types.StringValue(other.Proto)
	m.SrcIP = types.StringValue(other.Src)
	return diags
}
