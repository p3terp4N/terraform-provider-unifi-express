package provider

import (
	"context"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/apgroup"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/device"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/dns"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/firewall"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/network"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/portal"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/radius"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/routing"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/settings"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/site"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/user"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	ProviderUsernameDescription = "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` environment variable."
	ProviderPasswordDescription = "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` environment variable."
	ProviderAPIURLDescription   = "URL of the UniFi Express controller API. Can be specified with the `UNIFI_API` environment variable. " +
		"Typically `https://<express-ip>:8443`. You should **NOT** supply the path (`/api`), the SDK will discover the appropriate paths."
	ProviderSiteDescription          = "The site in the UniFi Express controller this provider will manage. Can be specified with the `UNIFI_SITE` environment variable. Default: `default`"
	ProviderAllowInsecureDescription = "Skip verification of TLS certificates of API requests. You may need to set this to `true` " +
		"if you are using your local API without setting up a signed certificate. Can be specified with the " +
		"`UNIFI_INSECURE` environment variable."
)

func NewV2(version string) func() provider.Provider {
	return func() provider.Provider {
		return &unifiProvider{
			version: version,
		}
	}
}

var (
	_ provider.Provider = &unifiProvider{}
)

type unifiProvider struct {
	version string
}

type unifiProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIUrl   types.String `tfsdk:"api_url"`
	Site     types.String `tfsdk:"site"`
	Insecure types.Bool   `tfsdk:"allow_insecure"`
}

func (p *unifiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unifi"
	resp.Version = p.version
}

func (p *unifiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: ProviderUsernameDescription,
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: ProviderPasswordDescription,
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: ProviderAPIURLDescription,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1), // workaround for `required: true`, because it fails on doc generation due to incorrectly detected difference between v1 and v2
					validators.HTTPSUrl(),
				},
				Optional: true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: ProviderSiteDescription,
				Optional:            true,
			},
			"allow_insecure": schema.BoolAttribute{
				MarkdownDescription: ProviderAllowInsecureDescription,
				Optional:            true,
			},
		},
	}
}

func (p *unifiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Unifi provider...")
	// Retrieve provider data from the configuration
	var cfg unifiProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if cfg.APIUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_url"),
			"Unknown UniFi Controller API URL",
			"The provider cannot create the UniFi Controller API client as there is an unknown configuration value "+
				"for the API endpoint. Either target apply the source of the value first, set the value statically in "+
				"the configuration, or use the UNIFI_API environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	// Check environment variables
	username := utils.GetAnyStringEnv("UNIFI_USERNAME")
	password := utils.GetAnyStringEnv("UNIFI_PASSWORD")
	apiUrl := utils.GetAnyStringEnv("UNIFI_API")
	site := utils.GetAnyStringEnv("UNIFI_SITE")
	insecure := utils.GetAnyBoolEnv("UNIFI_INSECURE")

	if !cfg.Username.IsNull() {
		username = cfg.Username.ValueString()
	}
	if !cfg.Password.IsNull() {
		password = cfg.Password.ValueString()
	}
	if !cfg.APIUrl.IsNull() {
		apiUrl = cfg.APIUrl.ValueString()
	}
	if !cfg.Site.IsNull() {
		site = cfg.Site.ValueString()
	}
	if !cfg.Insecure.IsNull() {
		insecure = cfg.Insecure.ValueBool()
	}
	if username == "" || password == "" {
		resp.Diagnostics.AddAttributeError(path.Root("username"), "Missing UniFi API credentials", "`username` and `password` must both be set")
	}
	if apiUrl == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_url"), "Missing UniFi API URL", "The `api_url` attribute must be set")
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if site == "" {
		site = "default" // set default site if not provided
	}
	c, err := base.NewClient(&base.ClientConfig{
		Username: username,
		Password: password,
		Url:      apiUrl,
		Site:     site,
		Insecure: insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create UniFi client", err.Error())
		return
	}
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *unifiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Device
		device.NewDeviceResource,
		device.NewPortProfileResource,
		// DNS
		dns.NewDnsRecordResource,
		dns.NewDynamicDNSResource,
		// Firewall
		firewall.NewFirewallGroupResource,
		firewall.NewFirewallRuleResource,
		firewall.NewFirewallZoneResource,
		firewall.NewFirewallZonePolicyResource,
		// Network
		network.NewNetworkResource,
		network.NewWLANResource,
		// Portal
		portal.NewPortalFileResource,
		// RADIUS
		radius.NewAccountResource,
		radius.NewRadiusProfileResource,
		// Routing
		routing.NewPortForwardResource,
		routing.NewStaticRouteResource,
		// Settings
		settings.NewAutoSpeedtestResource,
		settings.NewCountryResource,
		settings.NewDpiResource,
		settings.NewGuestAccessResource,
		settings.NewLocaleResource,
		settings.NewMagicSiteToSiteVpnResource,
		settings.NewNetworkOptimizationResource,
		settings.NewNtpResource,
		settings.NewRsyslogdResource,
		settings.NewSettingRadiusResource,
		settings.NewTeleportResource,
		settings.NewMgmtResource,
		settings.NewUsgResource,
		settings.NewUswResource,
		// Site
		site.NewSiteResource,
		// User
		user.NewUserGroupResource,
		user.NewUserResource,
	}
}

func (p *unifiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// AP Group
		apgroup.NewAPGroupDatasource,
		// Device
		device.NewPortProfileDatasource,
		// DNS
		dns.NewDnsRecordsDatasource,
		dns.NewDnsRecordDatasource,
		// Firewall
		firewall.NewFirewallZoneDatasource,
		// Network
		network.NewNetworkDatasource,
		// RADIUS
		radius.NewRadiusProfileDatasource,
		radius.NewAccountDatasource,
		// User
		user.NewUserDatasource,
		user.NewUserGroupDatasource,
	}
}
