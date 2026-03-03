package dns

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &dynamicDNSModel{}

type dynamicDNSModel struct {
	base.Model
	Interface types.String `tfsdk:"interface"`
	Service   types.String `tfsdk:"service"`
	HostName  types.String `tfsdk:"host_name"`
	Server    types.String `tfsdk:"server"`
	Login     types.String `tfsdk:"login"`
	Password  types.String `tfsdk:"password"`
}

func (m *dynamicDNSModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.DynamicDNS{
		ID:        m.ID.ValueString(),
		Interface: m.Interface.ValueString(),
		Service:   m.Service.ValueString(),
		HostName:  m.HostName.ValueString(),
		Server:    m.Server.ValueString(),
		Login:     m.Login.ValueString(),
		XPassword: m.Password.ValueString(),
	}, diag.Diagnostics{}
}

func (m *dynamicDNSModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.DynamicDNS)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.DynamicDNS")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Interface = types.StringValue(other.Interface)
	m.Service = types.StringValue(other.Service)
	m.HostName = types.StringValue(other.HostName)
	m.Server = ut.StringOrNull(other.Server)
	m.Login = ut.StringOrNull(other.Login)
	m.Password = ut.StringOrNull(other.XPassword)
	return diags
}
