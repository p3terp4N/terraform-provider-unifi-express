package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
)

var (
	wlanValidMinimumDataRate2g = []int{1000, 2000, 5500, 6000, 9000, 11000, 12000, 18000, 24000, 36000, 48000, 54000}
	wlanValidMinimumDataRate5g = []int{6000, 9000, 12000, 18000, 24000, 36000, 48000, 54000}
)

var (
	_ resource.Resource                = &wlanResource{}
	_ resource.ResourceWithConfigure   = &wlanResource{}
	_ resource.ResourceWithImportState = &wlanResource{}
	_ base.Resource                    = &wlanResource{}
)

type wlanResource struct {
	*base.GenericResource[*wlanModel]
}

func NewWLANResource() resource.Resource {
	return &wlanResource{
		GenericResource: base.NewGenericResource(
			"unifi_wlan",
			func() *wlanModel { return &wlanModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetWLAN(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateWLAN(ctx, site, body.(*unifi.WLAN))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateWLAN(ctx, site, body.(*unifi.WLAN))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteWLAN(ctx, site, id)
				},
			},
		),
	}
}

func int64OneOfValidator(allowed []int) validator.Int64 {
	values := make([]int64, len(allowed))
	for i, v := range allowed {
		values[i] = int64(v)
	}
	return int64validator.OneOf(values...)
}

func (r *wlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	allValid2g := append([]int{0}, wlanValidMinimumDataRate2g...)
	allValid5g := append([]int{0}, wlanValidMinimumDataRate5g...)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_wlan` resource manages wireless networks (SSIDs) on UniFi access points.\n\n" +
			"This resource allows you to create and manage WiFi networks with various security options including WPA2, WPA3, " +
			"and enterprise authentication. You can configure features such as guest policies, minimum data rates, band steering, " +
			"and scheduled availability.\n\n" +
			"Each WLAN can be customized with different security settings, VLAN assignments, and client options to meet specific " +
			"networking requirements.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID("The unique identifier of the wireless network in the UniFi controller."),
			"site": ut.SiteAttribute("The name of the UniFi site where the wireless network should be created. If not specified, the default site will be used."),
			"name": schema.StringAttribute{
				MarkdownDescription: "The SSID (network name) that will be broadcast by the access points. Must be between 1 and 32 characters long.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"user_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user group that defines the rate limiting and firewall rules for clients on this network.",
				Required:            true,
			},
			"security": schema.StringAttribute{
				MarkdownDescription: "The security protocol for the wireless network. Valid values are:\n" +
					"  * `wpapsk` - WPA Personal (PSK) with WPA2/WPA3 options\n" +
					"  * `wpaeap` - WPA Enterprise (802.1x)\n" +
					"  * `open` - Open network (no encryption)",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("wpapsk", "wpaeap", "open"),
				},
			},
			"wpa3_support": schema.BoolAttribute{
				MarkdownDescription: "Enable WPA3 security protocol. Requires security to be set to `wpapsk` and PMF mode to be enabled. WPA3 provides enhanced security features over WPA2.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"wpa3_transition": schema.BoolAttribute{
				MarkdownDescription: "Enable WPA3 transition mode, which allows both WPA2 and WPA3 clients to connect. This provides backward compatibility while gradually transitioning to WPA3." +
					" Requires security to be set to `wpapsk` and `wpa3_support` to be true.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"pmf_mode": schema.StringAttribute{
				MarkdownDescription: "Protected Management Frames (PMF) mode. It cannot be disabled if using WPA3. Valid values are:\n" +
					"  * `required` - All clients must support PMF (required for WPA3)\n" +
					"  * `optional` - Clients can optionally use PMF (recommended when transitioning from WPA2 to WPA3)\n" +
					"  * `disabled` - PMF is disabled (not compatible with WPA3)",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("disabled"),
				Validators: []validator.String{
					stringvalidator.OneOf("required", "optional", "disabled"),
				},
			},
			"passphrase": schema.StringAttribute{
				MarkdownDescription: "The WPA pre-shared key (password) for the network. Required when security is not set to `open`.",
				Optional:            true,
				Sensitive:           true,
			},
			"hide_ssid": schema.BoolAttribute{
				MarkdownDescription: "When enabled, the access points will not broadcast the network name (SSID). Clients will need to manually enter the SSID to connect.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"is_guest": schema.BoolAttribute{
				MarkdownDescription: "Mark this as a guest network. Guest networks are isolated from other networks and can have special restrictions like captive portals.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"multicast_enhance": schema.BoolAttribute{
				MarkdownDescription: "Enable multicast enhancement to convert multicast traffic to unicast for better reliability and performance, especially for applications like video streaming.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"mac_filter_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable MAC address filtering to control network access based on client MAC addresses. Works in conjunction with `mac_filter_list` and `mac_filter_policy`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"mac_filter_list": schema.SetAttribute{
				MarkdownDescription: "List of MAC addresses to filter in XX:XX:XX:XX:XX:XX format. Only applied when `mac_filter_enabled` is true. MAC addresses are case-insensitive.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"mac_filter_policy": schema.StringAttribute{
				MarkdownDescription: "MAC address filter policy. Valid values are:\n" +
					"  * `allow` - Only allow listed MAC addresses\n" +
					"  * `deny` - Block listed MAC addresses",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("deny"),
				Validators: []validator.String{
					stringvalidator.OneOf("allow", "deny"),
				},
			},
			"radius_profile_id": schema.StringAttribute{
				MarkdownDescription: "ID of the RADIUS profile to use for WPA Enterprise authentication (when security is 'wpaeap'). Reference existing profiles using the `unifi_radius_profile` data source.",
				Optional:            true,
			},
			"no2ghz_oui": schema.BoolAttribute{
				MarkdownDescription: "When enabled, devices from specific manufacturers (identified by their OUI - Organizationally Unique Identifier) " +
					"will be prevented from connecting on 2.4GHz and forced to use 5GHz. This improves overall network performance by " +
					"ensuring capable devices use the less congested 5GHz band. Common examples include newer smartphones and laptops.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"l2_isolation": schema.BoolAttribute{
				MarkdownDescription: "Isolates wireless clients from each other at layer 2 (ethernet) level. When enabled, devices on this WLAN " +
					"cannot communicate directly with each other, improving security especially for guest networks or IoT devices. " +
					"Each client can only communicate with the gateway/router.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"proxy_arp": schema.BoolAttribute{
				MarkdownDescription: "Enable ARP proxy on this WLAN. When enabled, the UniFi controller will respond to ARP requests on behalf " +
					"of clients, reducing broadcast traffic and potentially improving network performance. This is particularly useful " +
					"in high-density wireless environments.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"bss_transition": schema.BoolAttribute{
				MarkdownDescription: "Enable BSS Transition Management to help clients roam between APs more efficiently.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"uapsd": schema.BoolAttribute{
				MarkdownDescription: "Enable Unscheduled Automatic Power Save Delivery to improve battery life for mobile devices.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"fast_roaming_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable 802.11r Fast BSS Transition for seamless roaming between APs. Requires client device support.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"minimum_data_rate_2g_kbps": schema.Int64Attribute{
				MarkdownDescription: "Minimum data rate for 2.4GHz devices in Kbps. Use `0` to disable. Valid values: " +
					utils.MarkdownValueListInt(wlanValidMinimumDataRate2g),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64OneOfValidator(allValid2g),
				},
			},
			"minimum_data_rate_5g_kbps": schema.Int64Attribute{
				MarkdownDescription: "Minimum data rate for 5GHz devices in Kbps. Use `0` to disable. Valid values: " +
					utils.MarkdownValueListInt(wlanValidMinimumDataRate5g),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64OneOfValidator(allValid5g),
				},
			},
			"wlan_band": schema.StringAttribute{
				MarkdownDescription: "Radio band selection. Valid values:\n" +
					"  * `both` - Both 2.4GHz and 5GHz (default)\n" +
					"  * `2g` - 2.4GHz only\n" +
					"  * `5g` - 5GHz only",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("both"),
				Validators: []validator.String{
					stringvalidator.OneOf("2g", "5g", "both"),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "ID of the network (VLAN) for this SSID. Used to assign the WLAN to a specific network segment.",
				Optional:            true,
			},
			"ap_group_ids": schema.SetAttribute{
				MarkdownDescription: "IDs of the AP groups that should broadcast this SSID. Used to control which access points broadcast this network.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"schedule": schema.ListNestedBlock{
				MarkdownDescription: "Time-based access control configuration for the wireless network. Allows automatic enabling/disabling of the network on specified schedules.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"day_of_week": schema.StringAttribute{
							MarkdownDescription: "Day of week. Valid values: `sun`, `mon`, `tue`, `wed`, `thu`, `fri`, `sat`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("sun", "mon", "tue", "wed", "thu", "fri", "sat"),
							},
						},
						"start_hour": schema.Int64Attribute{
							MarkdownDescription: "Start hour in 24-hour format (0-23).",
							Required:            true,
							Validators: []validator.Int64{
								int64validator.Between(0, 23),
							},
						},
						"start_minute": schema.Int64Attribute{
							MarkdownDescription: "Start minute (0-59).",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
							Validators: []validator.Int64{
								int64validator.Between(0, 59),
							},
						},
						"duration": schema.Int64Attribute{
							MarkdownDescription: "Duration in minutes that the network should remain active.",
							Required:            true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Friendly name for this schedule block (e.g., 'Business Hours', 'Weekend Access').",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}
