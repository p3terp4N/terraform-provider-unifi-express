package device

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
)

var (
	_ resource.Resource                = &portProfileResource{}
	_ resource.ResourceWithConfigure   = &portProfileResource{}
	_ resource.ResourceWithImportState = &portProfileResource{}
	_ base.Resource                    = &portProfileResource{}
)

type portProfileResource struct {
	*base.GenericResource[*portProfileModel]
}

func NewPortProfileResource() resource.Resource {
	return &portProfileResource{
		GenericResource: base.NewGenericResource(
			"unifi_port_profile",
			func() *portProfileModel { return &portProfileModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetPortProfile(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.CreatePortProfile(ctx, site, model.(*unifi.PortProfile))
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.UpdatePortProfile(ctx, site, model.(*unifi.PortProfile))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeletePortProfile(ctx, site, id)
				},
			},
		),
	}
}

func (r *portProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_port_profile` resource manages port profiles that can be applied to UniFi switch ports.\n\n" +
			"Port profiles define a collection of settings that can be applied to one or more switch ports, including:\n" +
			"  * Network and VLAN settings\n" +
			"  * Port speed and duplex settings\n" +
			"  * Security features like 802.1X authentication and port isolation\n" +
			"  * Rate limiting and QoS settings\n" +
			"  * Network protocols like LLDP and STP\n\n" +
			"Creating port profiles allows for consistent configuration across multiple switch ports and easier management of port settings.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"autoneg": schema.BoolAttribute{
				MarkdownDescription: "Enable automatic negotiation of port speed and duplex settings. When enabled, this overrides manual speed and duplex settings. Recommended for most use cases.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"dot1x_ctrl": schema.StringAttribute{
				MarkdownDescription: "802.1X port-based network access control (PNAC) mode. Valid values are:\n" +
					"  * `force_authorized` - Port allows all traffic, no authentication required (default)\n" +
					"  * `force_unauthorized` - Port blocks all traffic regardless of authentication\n" +
					"  * `auto` - Standard 802.1X authentication required before port access is granted\n" +
					"  * `mac_based` - Authentication based on client MAC address, useful for devices that don't support 802.1X\n" +
					"  * `multi_host` - Allows multiple devices after first successful authentication, common in VoIP phone setups\n\n" +
					"Use 'auto' for highest security, 'mac_based' for legacy devices, and 'multi_host' when daisy-chaining devices.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("force_authorized"),
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "force_authorized", "force_unauthorized", "mac_based", "multi_host"),
				},
			},
			"dot1x_idle_timeout": schema.Int64Attribute{
				MarkdownDescription: "The number of seconds before an inactive authenticated MAC address is removed when using MAC-based 802.1X control. Range: 0-65535 seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"egress_rate_limit_kbps": schema.Int64Attribute{
				MarkdownDescription: "The maximum outbound bandwidth allowed on the port in kilobits per second. Range: 64-9999999 kbps. Only applied when egress_rate_limit_kbps_enabled is true.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(64, 9999999),
				},
			},
			"egress_rate_limit_kbps_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable outbound bandwidth rate limiting on the port. When enabled, traffic will be limited to the rate specified in egress_rate_limit_kbps.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"excluded_network_ids": schema.SetAttribute{
				MarkdownDescription: "List of network IDs to exclude when forward is set to 'customize'. This allows you to prevent specific networks from being accessible on ports using this profile.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"forward": schema.StringAttribute{
				MarkdownDescription: "VLAN forwarding mode for the port. Valid values are:\n" +
					"  * `all` - Forward all VLANs (trunk port)\n" +
					"  * `native` - Only forward untagged traffic (access port)\n" +
					"  * `customize` - Forward selected VLANs (use with `excluded_network_ids`)\n" +
					"  * `disabled` - Disable VLAN forwarding\n\n" +
					"Examples:\n" +
					"  * Use 'all' for uplink ports or connections to VLAN-aware devices\n" +
					"  * Use 'native' for end-user devices or simple network connections\n" +
					"  * Use 'customize' to create a selective trunk port (e.g., for a server needing access to specific VLANs)",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("native"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "native", "customize", "disabled"),
				},
			},
			"full_duplex": schema.BoolAttribute{
				MarkdownDescription: "Enable full-duplex mode when auto-negotiation is disabled. Full duplex allows simultaneous two-way communication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"isolation": schema.BoolAttribute{
				MarkdownDescription: "Enable port isolation. When enabled, devices connected to ports with this profile cannot communicate with each other, providing enhanced security.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"lldpmed_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable Link Layer Discovery Protocol-Media Endpoint Discovery (LLDP-MED). This allows for automatic discovery and configuration of devices like VoIP phones.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"lldpmed_notify_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable LLDP-MED topology change notifications. When enabled:\n" +
					"* Network devices will be notified of topology changes\n" +
					"* Useful for VoIP phones and other LLDP-MED capable devices\n" +
					"* Helps maintain accurate network topology information\n" +
					"* Facilitates faster device configuration and provisioning",
				Optional: true,
			},
			"native_networkconf_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network to use as the native (untagged) network on ports using this profile. " +
					"This is typically used for:\n" +
					"* Access ports where devices need untagged access\n" +
					"* Trunk ports to specify the native VLAN\n" +
					"* Management networks for network devices",
				Optional: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A descriptive name for the port profile. Examples:\n" +
					"* 'AP-Trunk-Port' - For access point uplinks\n" +
					"* 'VoIP-Phone-Port' - For VoIP phone connections\n" +
					"* 'User-Access-Port' - For standard user connections\n" +
					"* 'IoT-Device-Port' - For IoT device connections",
				Optional: true,
			},
			"op_mode": schema.StringAttribute{
				MarkdownDescription: "The operation mode for the port profile. Can only be `switch`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("switch"),
				Validators: []validator.String{
					stringvalidator.OneOf("switch"),
				},
			},
			"poe_mode": schema.StringAttribute{
				MarkdownDescription: "The POE mode for the port profile. Can be one of `auto`, `passv24`, `passthrough` or `off`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "passv24", "passthrough", "off"),
				},
			},
			"port_security_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable MAC address-based port security. When enabled:\n" +
					"* Only devices with specified MAC addresses can connect\n" +
					"* Unauthorized devices will be blocked\n" +
					"* Provides protection against unauthorized network access\n" +
					"* Must be used with port_security_mac_address list",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"port_security_mac_address": schema.SetAttribute{
				MarkdownDescription: "List of allowed MAC addresses when port security is enabled. Each address should be:\n" +
					"* In standard format (e.g., 'aa:bb:cc:dd:ee:ff')\n" +
					"* Unique per device\n" +
					"* Verified to belong to authorized devices\n" +
					"Only effective when port_security_enabled is true.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"priority_queue1_level": schema.Int64Attribute{
				MarkdownDescription: "Priority queue 1 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* Low-priority background traffic\n" +
					"* Bulk data transfers\n" +
					"* Non-time-sensitive applications\n" +
					"Higher values give more bandwidth to this queue.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"priority_queue2_level": schema.Int64Attribute{
				MarkdownDescription: "Priority queue 2 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* Standard user traffic\n" +
					"* Web browsing and email\n" +
					"* General business applications\n" +
					"Higher values give more bandwidth to this queue.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"priority_queue3_level": schema.Int64Attribute{
				MarkdownDescription: "Priority queue 3 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* High-priority traffic\n" +
					"* Voice and video conferencing\n" +
					"* Time-sensitive applications\n" +
					"Higher values give more bandwidth to this queue.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"priority_queue4_level": schema.Int64Attribute{
				MarkdownDescription: "Priority queue 4 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* Highest priority traffic\n" +
					"* Critical real-time applications\n" +
					"* Emergency communications\n" +
					"Higher values give more bandwidth to this queue.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"speed": schema.Int64Attribute{
				MarkdownDescription: "Port speed in Mbps when auto-negotiation is disabled. Common values:\n" +
					"* 10 - 10 Mbps (legacy devices)\n" +
					"* 100 - 100 Mbps (Fast Ethernet)\n" +
					"* 1000 - 1 Gbps (Gigabit Ethernet)\n" +
					"* 2500 - 2.5 Gbps (Multi-Gigabit)\n" +
					"* 5000 - 5 Gbps (Multi-Gigabit)\n" +
					"* 10000 - 10 Gbps (10 Gigabit)\n" +
					"Only used when autoneg is false.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.OneOf(10, 100, 1000, 2500, 5000, 10000, 20000, 25000, 40000, 50000, 100000),
				},
			},
			"stormctrl_bcast_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable broadcast storm control. When enabled:\n" +
					"* Limits broadcast traffic to prevent network flooding\n" +
					"* Protects against broadcast storms\n" +
					"* Helps maintain network stability\n" +
					"Use with stormctrl_bcast_rate to set threshold.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"stormctrl_bcast_level": schema.Int64Attribute{
				MarkdownDescription: "The broadcast Storm Control level for the port profile. Can be between 0 and 100.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
					int64validator.ConflictsWith(path.MatchRoot("stormctrl_bcast_rate")),
				},
			},
			"stormctrl_bcast_rate": schema.Int64Attribute{
				MarkdownDescription: "Maximum broadcast traffic rate in packets per second (0 - 14880000). Used to:\n" +
					"* Control broadcast traffic levels\n" +
					"* Prevent network congestion\n" +
					"* Balance between necessary broadcasts and network protection\n" +
					"Only effective when `stormctrl_bcast_enabled` is true.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 14880000),
					int64validator.ConflictsWith(path.MatchRoot("stormctrl_bcast_level")),
				},
			},
			"stormctrl_mcast_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable multicast storm control. When enabled:\n" +
					"* Limits multicast traffic to prevent network flooding\n" +
					"* Important for networks with multicast applications\n" +
					"* Helps maintain quality of service\n" +
					"Use with `stormctrl_mcast_rate` to set threshold.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"stormctrl_mcast_level": schema.Int64Attribute{
				MarkdownDescription: "The multicast Storm Control level for the port profile. Can be between 0 and 100.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
					int64validator.ConflictsWith(path.MatchRoot("stormctrl_mcast_rate")),
				},
			},
			"stormctrl_mcast_rate": schema.Int64Attribute{
				MarkdownDescription: "Maximum multicast traffic rate in packets per second (0 - 14880000). Used to:\n" +
					"* Control multicast traffic levels\n" +
					"* Ensure bandwidth for critical multicast services\n" +
					"* Prevent multicast traffic from overwhelming the network\n" +
					"Only effective when stormctrl_mcast_enabled is true.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 14880000),
					int64validator.ConflictsWith(path.MatchRoot("stormctrl_mcast_level")),
				},
			},
			"stormctrl_type": schema.StringAttribute{
				MarkdownDescription: "The type of Storm Control to use for the port profile. Can be one of `level` or `rate`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("level", "rate"),
				},
			},
			"stormctrl_ucast_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable unknown unicast storm control. When enabled:\n" +
					"* Limits unknown unicast traffic to prevent flooding\n" +
					"* Protects against MAC spoofing attacks\n" +
					"* Helps maintain network performance\n" +
					"Use with stormctrl_ucast_rate to set threshold.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"stormctrl_ucast_level": schema.Int64Attribute{
				MarkdownDescription: "The unknown unicast Storm Control level for the port profile. Can be between 0 and 100.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
					int64validator.ConflictsWith(path.MatchRoot("stormctrl_ucast_rate")),
				},
			},
			"stormctrl_ucast_rate": schema.Int64Attribute{
				MarkdownDescription: "Maximum unknown unicast traffic rate in packets per second (0 - 14880000). Used to:\n" +
					"* Control unknown unicast traffic levels\n" +
					"* Prevent network saturation from unknown destinations\n" +
					"* Balance security with network usability\n" +
					"Only effective when stormctrl_ucast_enabled is true.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 14880000),
					int64validator.ConflictsWith(path.MatchRoot("stormctrl_ucast_level")),
				},
			},
			"stp_port_mode": schema.BoolAttribute{
				MarkdownDescription: "Spanning Tree Protocol (STP) configuration for the port. When enabled:\n" +
					"* Prevents network loops in switch-to-switch connections\n" +
					"* Provides automatic failover in redundant topologies\n" +
					"* Helps maintain network stability\n\n" +
					"Best practices:\n" +
					"* Enable on switch uplink ports\n" +
					"* Enable on ports connecting to other switches\n" +
					"* Can be disabled on end-device ports for faster initialization",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"tagged_vlan_mgmt": schema.StringAttribute{
				MarkdownDescription: "VLAN tagging behavior for the port. Valid values are:\n" +
					"* `auto` - Automatically handle VLAN tags (recommended)\n" +
					"    - Intelligently manages tagged and untagged traffic\n" +
					"    - Best for most deployments\n" +
					"* `block_all` - Block all VLAN tagged traffic\n" +
					"    - Use for security-sensitive ports\n" +
					"    - Prevents VLAN hopping attacks\n" +
					"* `custom` - Custom VLAN configuration\n" +
					"    - Manual control over VLAN behavior\n" +
					"    - For specific VLAN requirements",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "block_all", "custom"),
				},
			},
			"voice_networkconf_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network to use for Voice over IP (VoIP) traffic. Used for:\n" +
					"* Automatic VoIP VLAN configuration\n" +
					"* Voice traffic prioritization\n" +
					"* QoS settings for voice packets\n\n" +
					"Common scenarios:\n" +
					"* IP phone deployments with separate voice VLAN\n" +
					"* Unified communications systems\n" +
					"* Converged voice/data networks\n\n" +
					"Works in conjunction with LLDP-MED for automatic phone provisioning.",
				Optional: true,
			},
		},
	}
}
