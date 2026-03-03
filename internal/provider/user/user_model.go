package user

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &userModel{}

type userModel struct {
	base.Model

	MAC                types.String `tfsdk:"mac"`
	Name               types.String `tfsdk:"name"`
	UserGroupID        types.String `tfsdk:"user_group_id"`
	Note               types.String `tfsdk:"note"`
	FixedIP            types.String `tfsdk:"fixed_ip"`
	NetworkID          types.String `tfsdk:"network_id"`
	Blocked            types.Bool   `tfsdk:"blocked"`
	DevIdOverride      types.Int64  `tfsdk:"dev_id_override"`
	LocalDnsRecord     types.String `tfsdk:"local_dns_record"`
	AllowExisting      types.Bool   `tfsdk:"allow_existing"`
	SkipForgetOnDestroy types.Bool  `tfsdk:"skip_forget_on_destroy"`

	// Computed-only attributes
	Hostname types.String `tfsdk:"hostname"`
	IP       types.String `tfsdk:"ip"`
}

func (m *userModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	fixedIP := m.FixedIP.ValueString()
	localDnsRecord := m.LocalDnsRecord.ValueString()

	return &unifi.User{
		ID:                    m.ID.ValueString(),
		MAC:                   m.MAC.ValueString(),
		Name:                  m.Name.ValueString(),
		UserGroupID:           m.UserGroupID.ValueString(),
		Note:                  m.Note.ValueString(),
		FixedIP:               fixedIP,
		UseFixedIP:            fixedIP != "",
		LocalDNSRecord:        localDnsRecord,
		LocalDNSRecordEnabled: localDnsRecord != "",
		NetworkID:             m.NetworkID.ValueString(),
		Blocked:               m.Blocked.ValueBool(),
		DevIdOverride:         int(m.DevIdOverride.ValueInt64()),
	}, diag.Diagnostics{}
}

func (m *userModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.User)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.User")
		return diags
	}

	m.ID = types.StringValue(other.ID)
	m.MAC = types.StringValue(other.MAC)
	m.Name = types.StringValue(other.Name)
	m.UserGroupID = types.StringValue(other.UserGroupID)
	m.Note = types.StringValue(other.Note)
	m.NetworkID = types.StringValue(other.NetworkID)
	m.Blocked = types.BoolValue(other.Blocked)
	m.DevIdOverride = types.Int64Value(int64(other.DevIdOverride))
	m.Hostname = types.StringValue(other.Hostname)
	m.IP = types.StringValue(other.IP)

	if other.UseFixedIP {
		m.FixedIP = types.StringValue(other.FixedIP)
	} else {
		m.FixedIP = types.StringValue("")
	}

	if other.LocalDNSRecordEnabled {
		m.LocalDnsRecord = types.StringValue(other.LocalDNSRecord)
	} else {
		m.LocalDnsRecord = types.StringValue("")
	}

	return diags
}
