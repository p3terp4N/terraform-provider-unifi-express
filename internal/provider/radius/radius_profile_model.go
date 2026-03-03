package radius

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
)

var _ base.ResourceModel = &radiusProfileModel{}

// radiusServerModel represents a single RADIUS server entry (auth or acct).
type radiusServerModel struct {
	IP      types.String `tfsdk:"ip"`
	Port    types.Int64  `tfsdk:"port"`
	XSecret types.String `tfsdk:"xsecret"`
}

// radiusServerModelAttrTypes returns the attr.Type map for radiusServerModel,
// used when constructing types.List values with types.ObjectType element type.
func radiusServerModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip":      types.StringType,
		"port":    types.Int64Type,
		"xsecret": types.StringType,
	}
}

// radiusServerObjectType returns the types.ObjectType for the nested server model.
func radiusServerObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: radiusServerModelAttrTypes()}
}

type radiusProfileModel struct {
	base.Model
	Name                  types.String `tfsdk:"name"`
	AccountingEnabled     types.Bool   `tfsdk:"accounting_enabled"`
	InterimUpdateEnabled  types.Bool   `tfsdk:"interim_update_enabled"`
	InterimUpdateInterval types.Int64  `tfsdk:"interim_update_interval"`
	UseUsgAcctServer      types.Bool   `tfsdk:"use_usg_acct_server"`
	UseUsgAuthServer      types.Bool   `tfsdk:"use_usg_auth_server"`
	VlanEnabled           types.Bool   `tfsdk:"vlan_enabled"`
	VlanWlanMode          types.String `tfsdk:"vlan_wlan_mode"`
	AuthServers           types.List   `tfsdk:"auth_server"`
	AcctServers           types.List   `tfsdk:"acct_server"`
}

func (m *radiusProfileModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	profile := &unifi.RADIUSProfile{
		ID:                    m.ID.ValueString(),
		Name:                  m.Name.ValueString(),
		AccountingEnabled:     m.AccountingEnabled.ValueBool(),
		InterimUpdateEnabled:  m.InterimUpdateEnabled.ValueBool(),
		InterimUpdateInterval: int(m.InterimUpdateInterval.ValueInt64()),
		UseUsgAcctServer:      m.UseUsgAcctServer.ValueBool(),
		UseUsgAuthServer:      m.UseUsgAuthServer.ValueBool(),
		VLANEnabled:           m.VlanEnabled.ValueBool(),
		VLANWLANMode:          m.VlanWlanMode.ValueString(),
	}

	// Convert auth servers
	if !m.AuthServers.IsNull() && !m.AuthServers.IsUnknown() {
		var authServerModels []radiusServerModel
		diags.Append(m.AuthServers.ElementsAs(ctx, &authServerModels, false)...)
		if diags.HasError() {
			return nil, diags
		}
		authServers := make([]unifi.RADIUSProfileAuthServers, 0, len(authServerModels))
		for _, s := range authServerModels {
			authServers = append(authServers, unifi.RADIUSProfileAuthServers{
				IP:      s.IP.ValueString(),
				Port:    int(s.Port.ValueInt64()),
				XSecret: s.XSecret.ValueString(),
			})
		}
		profile.AuthServers = authServers
	}

	// Convert acct servers
	if !m.AcctServers.IsNull() && !m.AcctServers.IsUnknown() {
		var acctServerModels []radiusServerModel
		diags.Append(m.AcctServers.ElementsAs(ctx, &acctServerModels, false)...)
		if diags.HasError() {
			return nil, diags
		}
		acctServers := make([]unifi.RADIUSProfileAcctServers, 0, len(acctServerModels))
		for _, s := range acctServerModels {
			acctServers = append(acctServers, unifi.RADIUSProfileAcctServers{
				IP:      s.IP.ValueString(),
				Port:    int(s.Port.ValueInt64()),
				XSecret: s.XSecret.ValueString(),
			})
		}
		profile.AcctServers = acctServers
	}

	return profile, diags
}

func (m *radiusProfileModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.RADIUSProfile)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.RADIUSProfile")
		return diags
	}

	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.AccountingEnabled = types.BoolValue(other.AccountingEnabled)
	m.InterimUpdateEnabled = types.BoolValue(other.InterimUpdateEnabled)
	m.InterimUpdateInterval = types.Int64Value(int64(other.InterimUpdateInterval))
	m.UseUsgAcctServer = types.BoolValue(other.UseUsgAcctServer)
	m.UseUsgAuthServer = types.BoolValue(other.UseUsgAuthServer)
	m.VlanEnabled = types.BoolValue(other.VLANEnabled)
	m.VlanWlanMode = types.StringValue(other.VLANWLANMode)

	// Convert auth servers from unifi model to Framework types.List
	if len(other.AuthServers) > 0 {
		authServers := make([]radiusServerModel, 0, len(other.AuthServers))
		for _, s := range other.AuthServers {
			authServers = append(authServers, radiusServerModel{
				IP:      types.StringValue(s.IP),
				Port:    types.Int64Value(int64(s.Port)),
				XSecret: types.StringValue(s.XSecret),
			})
		}
		authList, d := types.ListValueFrom(ctx, radiusServerObjectType(), authServers)
		diags.Append(d...)
		if !diags.HasError() {
			m.AuthServers = authList
		}
	} else {
		m.AuthServers = types.ListNull(radiusServerObjectType())
	}

	// Convert acct servers from unifi model to Framework types.List
	if len(other.AcctServers) > 0 {
		acctServers := make([]radiusServerModel, 0, len(other.AcctServers))
		for _, s := range other.AcctServers {
			acctServers = append(acctServers, radiusServerModel{
				IP:      types.StringValue(s.IP),
				Port:    types.Int64Value(int64(s.Port)),
				XSecret: types.StringValue(s.XSecret),
			})
		}
		acctList, d := types.ListValueFrom(ctx, radiusServerObjectType(), acctServers)
		diags.Append(d...)
		if !diags.HasError() {
			m.AcctServers = acctList
		}
	} else {
		m.AcctServers = types.ListNull(radiusServerObjectType())
	}

	return diags
}
