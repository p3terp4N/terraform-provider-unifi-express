package device

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/validators"
)

var (
	_ resource.Resource                = &deviceResource{}
	_ resource.ResourceWithConfigure   = &deviceResource{}
	_ resource.ResourceWithImportState = &deviceResource{}
	_ base.Resource                    = &deviceResource{}
)

type deviceResource struct {
	*base.GenericResource[*deviceModel]
}

func NewDeviceResource() resource.Resource {
	return &deviceResource{
		GenericResource: base.NewGenericResource(
			"unifi_device",
			func() *deviceModel { return &deviceModel{} },
			base.ResourceFunctions{
				// Read is used by the generic ImportState; individual CRUD is overridden below.
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetDevice(ctx, site, id)
				},
				Create: nil,
				Update: nil,
				Delete: nil,
			},
		),
	}
}

func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_device` resource manages UniFi network devices such as access points, switches, gateways, etc.\n\n" +
			"Devices must first be adopted by the UniFi controller before they can be managed through Terraform. " +
			"This resource cannot create new devices, but instead allows you to manage existing devices that have already been adopted. " +
			"The recommended approach is to adopt devices through the UniFi controller UI first, then import them into Terraform using the device's MAC address.\n\n" +
			"This resource supports managing device names, port configurations, and other device-specific settings.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID("The unique identifier of the device in the UniFi controller."),
			"site": ut.SiteAttribute("The name of the UniFi site where the device is located. If not specified, the default site will be used."),
			"mac": schema.StringAttribute{
				MarkdownDescription: "The MAC address of the device in standard format (e.g., 'aa:bb:cc:dd:ee:ff'). " +
					"This is used to identify and manage specific devices that have already been adopted by the controller.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					ut.MACNormalization(),
				},
				Validators: []validator.String{
					validators.Mac,
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the device that will be displayed in the UniFi controller UI. Examples:\n" +
					"* 'Office-AP-1' for an access point\n" +
					"* 'Core-Switch-01' for a switch\n" +
					"* 'Main-Gateway' for a gateway\n" +
					"Choose descriptive names that indicate location and purpose.",
				Optional: true,
				Computed: true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the device is administratively disabled. When true, the device will not forward traffic or provide services.",
				Computed:            true,
			},
			"allow_adoption": schema.BoolAttribute{
				MarkdownDescription: "Whether to automatically adopt the device when creating this resource. When true:\n" +
					"* The controller will attempt to adopt the device\n" +
					"* Device must be in a pending adoption state\n" +
					"* Device must be accessible on the network\n" +
					"Set to false if you want to manage adoption manually.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"forget_on_destroy": schema.BoolAttribute{
				MarkdownDescription: "Whether to forget (un-adopt) the device when this resource is destroyed. When true:\n" +
					"* The device will be removed from the controller\n" +
					"* The device will need to be readopted to be managed again\n" +
					"* Device configuration will be reset\n" +
					"Set to false to keep the device adopted when removing from Terraform management.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"port_override": schema.SetNestedAttribute{
				MarkdownDescription: "A set of port-specific configuration overrides for UniFi switches. This allows you to customize individual port settings such as:\n" +
					"  * Port names and labels for easy identification\n" +
					"  * Port profiles for VLAN and security settings\n" +
					"  * Operating modes for special functions\n\n" +
					"Common use cases include:\n" +
					"  * Setting up trunk ports for inter-switch connections\n" +
					"  * Configuring PoE settings for powered devices\n" +
					"  * Creating mirrored ports for network monitoring\n" +
					"  * Setting up link aggregation between switches or servers",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"number": schema.Int64Attribute{
							MarkdownDescription: "The physical port number on the switch to configure.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "A friendly name for the port that will be displayed in the UniFi controller UI. Examples:\n" +
								"  * 'Uplink to Core Switch'\n" +
								"  * 'Conference Room AP'\n" +
								"  * 'Server LACP Group 1'\n" +
								"  * 'VoIP Phone Port'",
							Optional: true,
						},
						"port_profile_id": schema.StringAttribute{
							MarkdownDescription: "The ID of a pre-configured port profile to apply to this port. Port profiles define settings like VLANs, PoE, and other port-specific configurations.",
							Optional:            true,
						},
						"op_mode": schema.StringAttribute{
							MarkdownDescription: "The operating mode of the port. Valid values are:\n" +
								"  * `switch` - Normal switching mode (default)\n" +
								"    - Standard port operation for connecting devices\n" +
								"    - Supports VLANs and all standard switching features\n" +
								"  * `mirror` - Port mirroring for traffic analysis\n" +
								"    - Copies traffic from other ports for monitoring\n" +
								"    - Useful for network troubleshooting and security\n" +
								"  * `aggregate` - Link aggregation/bonding mode\n" +
								"    - Combines multiple ports for increased bandwidth\n" +
								"    - Used for switch uplinks or high-bandwidth servers",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("switch"),
							Validators: []validator.String{
								stringvalidator.OneOf("switch", "mirror", "aggregate"),
							},
							PlanModifiers: []planmodifier.String{
								opModeSuppressModifier{},
							},
						},
						"poe_mode": schema.StringAttribute{
							MarkdownDescription: "The Power over Ethernet (PoE) mode for the port. Valid values are:\n" +
								"* `auto` - Automatically detect and power PoE devices (recommended)\n" +
								"* `pasv24` - Passive 24V PoE\n" +
								"* `passthrough` - PoE passthrough mode\n" +
								"* `off` - Disable PoE on the port",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("auto", "pasv24", "passthrough", "off"),
							},
						},
						"aggregate_num_ports": schema.Int64Attribute{
							MarkdownDescription: "The number of ports to include in a link aggregation group (LAG). Valid range: 2-8 ports.",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(2, 8),
							},
						},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Plan modifiers that replicate V1 DiffSuppressFunc behavior
// ---------------------------------------------------------------------------

// opModeSuppressModifier suppresses diff when state is "" and plan is "switch"
// (API returns empty string for the default "switch" mode).
type opModeSuppressModifier struct{}

func (m opModeSuppressModifier) Description(_ context.Context) string {
	return "Suppresses diff when state is empty and plan is 'switch' (the default)."
}

func (m opModeSuppressModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m opModeSuppressModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.ValueString() == "" && req.PlanValue.ValueString() == "switch" {
		resp.PlanValue = req.StateValue
	}
}

// ---------------------------------------------------------------------------
// Custom CRUD
// ---------------------------------------------------------------------------

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.GetClient()
	site := client.ResolveSite(&plan)

	mac := plan.MAC.ValueString()
	if mac == "" {
		resp.Diagnostics.AddError("MAC address required", "No MAC address specified, please import the device using terraform import")
		return
	}
	mac = utils.CleanMAC(mac)

	device, err := client.GetDeviceByMAC(ctx, site, mac)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up device", err.Error())
		return
	}
	if device == nil {
		resp.Diagnostics.AddError("Device not found", fmt.Sprintf("Device not found using MAC %q", mac))
		return
	}

	if !device.Adopted {
		if !plan.AllowAdoption.ValueBool() {
			resp.Diagnostics.AddError("Device not adopted", "Device must be adopted before it can be managed")
			return
		}

		tflog.Debug(ctx, "Adopting device", map[string]interface{}{"mac": mac})
		err := client.AdoptDevice(ctx, site, mac)
		if err != nil {
			resp.Diagnostics.AddError("Error adopting device", err.Error())
			return
		}

		device, err = waitForDeviceState(ctx, client, site, mac, unifi.DeviceStateConnected,
			[]unifi.DeviceState{unifi.DeviceStateAdopting, unifi.DeviceStatePending, unifi.DeviceStateProvisioning, unifi.DeviceStateUpgrading},
			2*time.Minute)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for device adoption", err.Error())
			return
		}
	}

	plan.SetID(device.ID)
	plan.SetSite(site)

	// Apply configuration via Update
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	unifiDevice := body.(*unifi.Device)
	unifiDevice.ID = device.ID
	unifiDevice.SiteID = site

	updated, err := client.UpdateDevice(ctx, site, unifiDevice)
	if err != nil {
		resp.Diagnostics.AddError("Error updating device after adoption", err.Error())
		return
	}

	_, err = waitForDeviceState(ctx, client, site, mac, unifi.DeviceStateConnected,
		[]unifi.DeviceState{unifi.DeviceStateAdopting, unifi.DeviceStateProvisioning},
		1*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for device provisioning", err.Error())
		return
	}

	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.GetClient()
	site := client.ResolveSite(&state)

	device, err := client.GetDevice(ctx, site, state.GetID())
	if errors.Is(err, unifi.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading device", err.Error())
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, device)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.GetClient()
	site := client.ResolveSite(&plan)

	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	unifiDevice := body.(*unifi.Device)
	unifiDevice.ID = state.GetID()
	unifiDevice.SiteID = site

	updated, err := client.UpdateDevice(ctx, site, unifiDevice)
	if err != nil {
		resp.Diagnostics.AddError("Error updating device", err.Error())
		return
	}

	mac := plan.MAC.ValueString()
	if mac == "" {
		mac = state.MAC.ValueString()
	}

	_, err = waitForDeviceState(ctx, client, site, mac, unifi.DeviceStateConnected,
		[]unifi.DeviceState{unifi.DeviceStateAdopting, unifi.DeviceStateProvisioning},
		1*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for device provisioning", err.Error())
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, updated)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.ForgetOnDestoy.ValueBool() {
		return
	}

	client := r.GetClient()
	site := client.ResolveSite(&state)
	mac := state.MAC.ValueString()

	// Retry ForgetDevice in a loop when the device is busy
	timeout := time.Now().Add(1 * time.Minute)
	for {
		err := client.ForgetDevice(ctx, site, mac)
		if err == nil {
			break
		}
		if utils.IsServerErrorContains(err, "api.err.DeviceBusy") {
			if time.Now().After(timeout) {
				resp.Diagnostics.AddError("Timeout forgetting device", "Device remained busy for too long: "+err.Error())
				return
			}
			tflog.Debug(ctx, "Device busy, retrying ForgetDevice", map[string]interface{}{"mac": mac})
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Diagnostics.AddError("Error forgetting device", err.Error())
		return
	}

	// Wait for device to reach pending state or disappear
	_, err := waitForDeviceState(ctx, client, site, mac, unifi.DeviceStatePending,
		[]unifi.DeviceState{unifi.DeviceStateConnected, unifi.DeviceStateDeleting},
		1*time.Minute)
	// ErrNotFound is expected — the device may disappear entirely after being forgotten
	if err != nil && !errors.Is(err, unifi.ErrNotFound) {
		resp.Diagnostics.AddError("Error waiting for device forget", err.Error())
		return
	}
}

func (r *deviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "Expected configured client. Please report this issue to the provider developers.")
		return
	}

	id := req.ID
	site := client.Site

	// Support site:mac and site:id format.
	// A MAC has 5 colons, so "site:mac" has 6 colons total, while "site:id" has 1 colon.
	if colons := strings.Count(id, ":"); colons == 1 || colons == 6 {
		parts := strings.SplitN(id, ":", 2)
		site = parts[0]
		id = parts[1]
	}

	// If the id looks like a MAC address, resolve it to the device ID.
	if utils.MacAddressRegexp.MatchString(id) {
		mac := utils.CleanMAC(id)
		device, err := client.GetDeviceByMAC(ctx, site, mac)
		if err != nil {
			resp.Diagnostics.AddError("Error looking up device by MAC", err.Error())
			return
		}
		if device == nil {
			resp.Diagnostics.AddError("Device not found", fmt.Sprintf("No device found with MAC %q", mac))
			return
		}
		id = device.ID
	}

	// Read the device to populate state
	state := &deviceModel{}
	state.SetID(id)
	state.SetSite(site)

	device, err := client.GetDevice(ctx, site, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading device during import", err.Error())
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, device)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	// Set defaults for meta-attributes that are not returned by the API
	state.AllowAdoption = types.BoolValue(true)
	state.ForgetOnDestoy = types.BoolValue(true)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// Device state polling helper
// ---------------------------------------------------------------------------

// waitForDeviceState polls GetDeviceByMAC until the device reaches the target
// state or the timeout elapses. It treats pendingStates (plus Unknown) as
// transient states that will be retried. Returns the device once it reaches
// the target state, or an error on timeout / unexpected state.
func waitForDeviceState(
	ctx context.Context,
	client *base.Client,
	site, mac string,
	targetState unifi.DeviceState,
	pendingStates []unifi.DeviceState,
	timeout time.Duration,
) (*unifi.Device, error) {
	// Build a set of pending states (always include Unknown).
	pending := map[unifi.DeviceState]bool{unifi.DeviceStateUnknown: true}
	for _, s := range pendingStates {
		pending[s] = true
	}

	deadline := time.Now().Add(timeout)
	notFoundCount := 0
	const maxNotFound = 30

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for device %s to reach state %s", mac, targetState.String())
		}

		device, err := client.GetDeviceByMAC(ctx, site, mac)

		// Handle transient not-found / unknown-device errors that occur
		// briefly after a forget operation.
		if errors.Is(err, unifi.ErrNotFound) {
			err = nil
		}
		if err != nil && strings.Contains(err.Error(), "api.err.UnknownDevice") {
			err = nil
		}

		if err != nil {
			return nil, err
		}

		if device == nil {
			notFoundCount++
			if notFoundCount >= maxNotFound {
				return nil, unifi.ErrNotFound
			}
			tflog.Debug(ctx, "Device not found yet, waiting", map[string]interface{}{"mac": mac, "attempt": notFoundCount})
			time.Sleep(2 * time.Second)
			continue
		}

		notFoundCount = 0

		if device.State == targetState {
			return device, nil
		}

		if !pending[device.State] {
			return device, fmt.Errorf("device %s reached unexpected state %s (expected %s)", mac, device.State.String(), targetState.String())
		}

		tflog.Debug(ctx, "Device in transient state, waiting", map[string]interface{}{
			"mac":           mac,
			"current_state": device.State.String(),
			"target_state":  targetState.String(),
		})
		time.Sleep(2 * time.Second)
	}
}
