package settings

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &settingRadiusModel{}

type settingRadiusModel struct {
	base.Model
	Enabled               types.Bool   `tfsdk:"enabled"`
	AccountingEnabled     types.Bool   `tfsdk:"accounting_enabled"`
	AccountingPort        types.Int64  `tfsdk:"accounting_port"`
	AuthPort              types.Int64  `tfsdk:"auth_port"`
	InterimUpdateInterval types.Int64  `tfsdk:"interim_update_interval"`
	TunneledReply         types.Bool   `tfsdk:"tunneled_reply"`
	Secret                types.String `tfsdk:"secret"`
}

func (m *settingRadiusModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.SettingRadius{
		ID:                    m.ID.ValueString(),
		Enabled:               m.Enabled.ValueBool(),
		AccountingEnabled:     m.AccountingEnabled.ValueBool(),
		AcctPort:              int(m.AccountingPort.ValueInt64()),
		AuthPort:              int(m.AuthPort.ValueInt64()),
		InterimUpdateInterval: int(m.InterimUpdateInterval.ValueInt64()),
		TunneledReply:         m.TunneledReply.ValueBool(),
		XSecret:               m.Secret.ValueString(),
		ConfigureWholeNetwork: true,
	}, diag.Diagnostics{}
}

func (m *settingRadiusModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.SettingRadius)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.SettingRadius")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Enabled = types.BoolValue(other.Enabled)
	m.AccountingEnabled = types.BoolValue(other.AccountingEnabled)
	m.AccountingPort = types.Int64Value(int64(other.AcctPort))
	m.AuthPort = types.Int64Value(int64(other.AuthPort))
	m.InterimUpdateInterval = types.Int64Value(int64(other.InterimUpdateInterval))
	m.TunneledReply = types.BoolValue(other.TunneledReply)
	m.Secret = types.StringValue(other.XSecret)
	return diags
}
