package radius

import (
	"context"
	"fmt"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
)

var (
	_ resource.Resource                = &radiusProfileResource{}
	_ resource.ResourceWithConfigure   = &radiusProfileResource{}
	_ resource.ResourceWithImportState = &radiusProfileResource{}
	_ base.Resource                    = &radiusProfileResource{}
)

type radiusProfileResource struct {
	*base.GenericResource[*radiusProfileModel]
}

func NewRadiusProfileResource() resource.Resource {
	return &radiusProfileResource{
		GenericResource: base.NewGenericResource(
			"unifi_radius_profile",
			func() *radiusProfileModel { return &radiusProfileModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetRADIUSProfile(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.CreateRADIUSProfile(ctx, site, model.(*unifi.RADIUSProfile))
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.UpdateRADIUSProfile(ctx, site, model.(*unifi.RADIUSProfile))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteRADIUSProfile(ctx, site, id)
				},
			},
		),
	}
}

func (r *radiusProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider client is not configured")
		return
	}

	id := req.ID
	site := client.Site

	// Support site:id format
	if strings.Contains(id, ":") {
		parts := strings.SplitN(id, ":", 2)
		site = parts[0]
		id = parts[1]
	}

	// Support name=<name> lookup
	if strings.HasPrefix(id, "name=") {
		targetName := strings.TrimPrefix(id, "name=")
		resolvedID, err := getRadiusProfileIDByName(ctx, client.Client, targetName, site)
		if err != nil {
			resp.Diagnostics.AddError("Error importing RADIUS profile by name", err.Error())
			return
		}
		id = resolvedID
	}

	state := &radiusProfileModel{}
	state.SetID(id)
	state.SetSite(site)

	// Read the resource to populate state
	res, err := client.GetRADIUSProfile(ctx, site, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading RADIUS profile", err.Error())
		return
	}
	state.Merge(ctx, res)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *radiusProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	serverAttributes := map[string]schema.Attribute{
		"ip": schema.StringAttribute{
			MarkdownDescription: "The IPv4 address of the RADIUS server (e.g., '192.168.1.100'). Must be reachable from your UniFi network.",
			Required:            true,
		},
		"port": schema.Int64Attribute{
			MarkdownDescription: "The UDP port number where the RADIUS service is listening. Valid values are between 1 and 65535.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 65535),
			},
		},
		"xsecret": schema.StringAttribute{
			MarkdownDescription: "The shared secret key used to secure communication between the UniFi controller and the RADIUS server. " +
				"This must match the secret configured on your RADIUS server.",
			Required:  true,
			Sensitive: true,
		},
	}

	// Clone the server attributes for acct_server with different default port
	acctServerAttributes := make(map[string]schema.Attribute, len(serverAttributes))
	for k, v := range serverAttributes {
		acctServerAttributes[k] = v
	}
	// Override port default for auth servers (1812)
	authPortAttr := serverAttributes["port"].(schema.Int64Attribute)
	authPortAttr.Default = int64default.StaticInt64(1812)
	serverAttributes["port"] = authPortAttr

	// Override port default for acct servers (1813)
	acctPortAttr := acctServerAttributes["port"].(schema.Int64Attribute)
	acctPortAttr.Default = int64default.StaticInt64(1813)
	acctServerAttributes["port"] = acctPortAttr

	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_radius_profile` resource manages RADIUS authentication profiles for UniFi networks.\n\n" +
			"RADIUS (Remote Authentication Dial-In User Service) profiles enable enterprise-grade authentication and authorization for:\n" +
			"  * 802.1X network access control\n" +
			"  * WPA2/WPA3-Enterprise wireless networks\n" +
			"  * Dynamic VLAN assignment\n" +
			"  * User activity accounting\n\n" +
			"Each profile can be configured with:\n" +
			"  * Multiple authentication and accounting servers\n" +
			"  * VLAN assignment settings\n" +
			"  * Accounting update intervals",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID("The unique identifier of the RADIUS profile in the UniFi controller."),
			"site": ut.SiteAttribute("The name of the UniFi site where the RADIUS profile should be created. If not specified, the default site will be used."),
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the RADIUS profile to help identify its purpose (e.g., 'Corporate Users' or 'Guest Access').",
				Required:            true,
			},
			"accounting_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable RADIUS accounting to track user sessions, including login/logout times and data usage. Useful for billing and audit purposes.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"interim_update_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable periodic updates during active sessions. This allows tracking of ongoing session data like bandwidth usage.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"interim_update_interval": schema.Int64Attribute{
				MarkdownDescription: "The interval (in seconds) between interim updates when `interim_update_enabled` is true. Default is 3600 seconds (1 hour).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3600),
			},
			"use_usg_acct_server": schema.BoolAttribute{
				MarkdownDescription: "Use the controller as a RADIUS accounting server. This allows local accounting without an external RADIUS server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"use_usg_auth_server": schema.BoolAttribute{
				MarkdownDescription: "Use the controller as a RADIUS authentication server. This allows local authentication without an external RADIUS server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"vlan_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable VLAN assignment for wired clients based on RADIUS attributes. This allows network segmentation based on user authentication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"vlan_wlan_mode": schema.StringAttribute{
				MarkdownDescription: "VLAN assignment mode for wireless networks. Valid values are:\n" +
					"  * `disabled` - Do not use RADIUS-assigned VLANs\n" +
					"  * `optional` - Use RADIUS-assigned VLAN if provided\n" +
					"  * `required` - Require RADIUS-assigned VLAN for authentication to succeed",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.OneOf("", "disabled", "optional", "required"),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"auth_server": schema.ListNestedBlock{
				MarkdownDescription: "List of RADIUS authentication servers to use with this profile. Multiple servers provide failover - if the first " +
					"server is unreachable, the system will try the next server in the list.",
				NestedObject: schema.NestedBlockObject{
					Attributes: serverAttributes,
				},
			},
			"acct_server": schema.ListNestedBlock{
				MarkdownDescription: "List of RADIUS accounting servers to use with this profile. Accounting servers track session data like " +
					"connection time and data usage.",
				NestedObject: schema.NestedBlockObject{
					Attributes: acctServerAttributes,
				},
			},
		},
	}
}

func getRadiusProfileIDByName(ctx context.Context, client unifi.Client, profileName, site string) (string, error) {
	radiusProfiles, err := client.ListRADIUSProfile(ctx, site)
	if err != nil {
		return "", err
	}

	idMatchingName := ""
	allNames := []string{}
	for _, profile := range radiusProfiles {
		allNames = append(allNames, profile.Name)
		if profile.Name != profileName {
			continue
		}
		if idMatchingName != "" {
			return "", fmt.Errorf("found multiple RADIUS profiles with name '%s'", profileName)
		}
		idMatchingName = profile.ID
	}
	if idMatchingName == "" {
		return "", fmt.Errorf("found no RADIUS profile with name '%s', found: %s", profileName, strings.Join(allNames, ", "))
	}
	return idMatchingName, nil
}
