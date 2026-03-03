package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &userDatasource{}
	_ datasource.DataSourceWithConfigure = &userDatasource{}
	_ base.Resource                      = &userDatasource{}
)

type userDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type userDatasourceModel struct {
	base.Model
	MAC            types.String `tfsdk:"mac"`
	Name           types.String `tfsdk:"name"`
	UserGroupID    types.String `tfsdk:"user_group_id"`
	Note           types.String `tfsdk:"note"`
	FixedIP        types.String `tfsdk:"fixed_ip"`
	NetworkID      types.String `tfsdk:"network_id"`
	Blocked        types.Bool   `tfsdk:"blocked"`
	DevIdOverride  types.Int64  `tfsdk:"dev_id_override"`
	Hostname       types.String `tfsdk:"hostname"`
	IP             types.String `tfsdk:"ip"`
	LocalDnsRecord types.String `tfsdk:"local_dns_record"`
}

func NewUserDatasource() datasource.DataSource {
	return &userDatasource{}
}

func (d *userDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *userDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *userDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *userDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *userDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_user", req.ProviderTypeName)
}

func (d *userDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_user` retrieves properties of a user (or \"client\" in the UI) of the network by MAC address.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"mac": schema.StringAttribute{
				MarkdownDescription: "The MAC address of the user.",
				Required:            true,
				Validators: []validator.String{
					validators.Mac,
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user.",
				Computed:            true,
			},
			"user_group_id": schema.StringAttribute{
				MarkdownDescription: "The user group ID for the user.",
				Computed:            true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "A note with additional information for the user.",
				Computed:            true,
			},
			"fixed_ip": schema.StringAttribute{
				MarkdownDescription: "Fixed IPv4 address set for this user.",
				Computed:            true,
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The network ID for this user.",
				Computed:            true,
			},
			"blocked": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether this user is blocked from the network.",
				Computed:            true,
			},
			"dev_id_override": schema.Int64Attribute{
				MarkdownDescription: "Override the device fingerprint.",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname of the user.",
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address of the user.",
				Computed:            true,
			},
			"local_dns_record": schema.StringAttribute{
				MarkdownDescription: "The local DNS record for this user.",
				Computed:            true,
			},
		},
	}
}

func (d *userDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)
	mac := strings.ToLower(state.MAC.ValueString())

	macResp, err := d.client.GetUserByMAC(ctx, site, mac)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get user by MAC", err.Error())
		return
	}

	user, err := d.client.GetUser(ctx, site, macResp.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get user", err.Error())
		return
	}

	state.SetID(user.ID)
	state.SetSite(site)
	state.MAC = types.StringValue(user.MAC)
	state.Name = ut.StringOrNull(user.Name)
	state.UserGroupID = ut.StringOrNull(user.UserGroupID)
	state.Note = ut.StringOrNull(user.Note)
	state.NetworkID = ut.StringOrNull(user.NetworkID)
	state.Blocked = types.BoolValue(user.Blocked)
	state.DevIdOverride = types.Int64Value(int64(user.DevIdOverride))
	state.Hostname = ut.StringOrNull(user.Hostname)
	state.IP = ut.StringOrNull(macResp.IP)

	fixedIP := ""
	if user.UseFixedIP {
		fixedIP = user.FixedIP
	}
	state.FixedIP = ut.StringOrNull(fixedIP)

	localDnsRecord := ""
	if user.LocalDNSRecordEnabled {
		localDnsRecord = user.LocalDNSRecord
	}
	state.LocalDnsRecord = ut.StringOrNull(localDnsRecord)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
