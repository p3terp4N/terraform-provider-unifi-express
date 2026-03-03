package user

import (
	"context"
	"errors"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
	_ base.Resource                    = &userResource{}
)

type userResource struct {
	*base.GenericResource[*userModel]
}

func NewUserResource() resource.Resource {
	return &userResource{
		GenericResource: base.NewGenericResource(
			"unifi_user",
			func() *userModel { return &userModel{} },
			base.ResourceFunctions{
				// CRUD handlers are nil — all four operations are overridden
				// on userResource because this resource has custom logic for
				// allow_existing, block/unblock, dev_id_override fingerprint,
				// skip_forget_on_destroy, and the extra GetUserByMAC call.
				Read:   nil,
				Create: nil,
				Update: nil,
				Delete: nil,
			},
		),
	}
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_user` resource manages network clients in the UniFi controller, which are identified by their unique MAC addresses.\n\n" +
			"This resource allows you to manage:\n" +
			"  * Fixed IP assignments\n" +
			"  * User groups and network access\n" +
			"  * Network blocking and restrictions\n" +
			"  * Local DNS records\n\n" +
			"Important Notes:\n" +
			"  * Users are automatically created in the controller when devices connect to the network\n" +
			"  * By default, this resource can take over management of existing users (controlled by `allow_existing`)\n" +
			"  * Users can be 'forgotten' on destroy (controlled by `skip_forget_on_destroy`)\n\n" +
			"This resource is particularly useful for:\n" +
			"  * Managing static IP assignments\n" +
			"  * Implementing access control\n" +
			"  * Setting up local DNS records\n" +
			"  * Organizing devices into user groups",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"mac": schema.StringAttribute{
				MarkdownDescription: "The MAC address of the device/client. This is used as the unique identifier and cannot be changed " +
					"after creation. Must be a valid MAC address format (e.g., '00:11:22:33:44:55'). MAC addresses are case-insensitive.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					ut.MACNormalization(),
				},
				Validators: []validator.String{
					validators.Mac,
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the device/client. This helps identify the device in the UniFi interface " +
					"(eg. 'Living Room TV', 'John's Laptop').",
				Required: true,
			},
			"user_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user group this client belongs to. User groups can be used to apply common " +
					"settings and restrictions to multiple clients.",
				Optional: true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Additional information about the client that you want to record (e.g., 'Company asset tag #12345', " +
					"'Guest device - expires 2024-03-01').",
				Optional: true,
			},
			"fixed_ip": schema.StringAttribute{
				MarkdownDescription: "A static IPv4 address to assign to this client. Ensure this IP is within the client's network range " +
					"and not already assigned to another device.",
				Optional: true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network this client should be associated with. This is particularly important " +
					"when using VLANs or multiple networks.",
				Optional: true,
			},
			"blocked": schema.BoolAttribute{
				MarkdownDescription: "When true, this client will be blocked from accessing the network. Useful for temporarily " +
					"or permanently restricting network access for specific devices.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"dev_id_override": schema.Int64Attribute{
				MarkdownDescription: "Override the device fingerprint.",
				Optional:            true,
			},
			"local_dns_record": schema.StringAttribute{
				MarkdownDescription: "A local DNS hostname for this client. When set, other devices on the network can resolve " +
					"this name to the client's IP address (e.g., 'printer.local', 'nas.home.arpa'). Such DNS record is automatically added to controller's DNS records.",
				Optional: true,
			},

			// Meta-attributes that control Terraform UX, not sent to API
			"allow_existing": schema.BoolAttribute{
				MarkdownDescription: "Allow this resource to take over management of an existing user in the UniFi controller. When true:\n" +
					"  * The resource can manage users that were automatically created when devices connected\n" +
					"  * Existing settings will be overwritten with the values specified in this resource\n" +
					"  * If false, attempting to manage an existing user will result in an error\n\n" +
					"Use with caution as it can modify settings for devices already connected to your network.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"skip_forget_on_destroy": schema.BoolAttribute{
				MarkdownDescription: "When false (default), the client will be 'forgotten' by the controller when this resource is destroyed. " +
					"Set to true to keep the client's history in the controller after the resource is removed from Terraform.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},

			// Computed-only attributes
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname of the user.",
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address of the user.",
				Computed:            true,
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.GetClient()
	site := c.ResolveSite(&plan)

	body, diags := plan.AsUnifiModel(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	userReq := body.(*unifi.User)

	allowExisting := plan.AllowExisting.ValueBool()

	result, err := c.CreateUser(ctx, site, userReq)
	if err != nil {
		if !utils.IsServerErrorContains(err, "api.err.MacUsed") || !allowExisting {
			resp.Diagnostics.AddError("Error creating user", err.Error())
			return
		}

		// MAC in use — absorb the existing user
		mac := plan.MAC.ValueString()
		existing, err := c.GetUserByMAC(ctx, site, mac)
		if err != nil {
			resp.Diagnostics.AddError("Error looking up existing user by MAC", err.Error())
			return
		}

		userReq.ID = existing.ID
		userReq.SiteID = existing.SiteID

		result, err = c.UpdateUser(ctx, site, userReq)
		if err != nil {
			resp.Diagnostics.AddError("Error updating existing user", err.Error())
			return
		}
	}

	// Handle block/unblock
	if plan.Blocked.ValueBool() {
		err := c.BlockUserByMAC(ctx, site, plan.MAC.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error blocking user", err.Error())
			return
		}
	}

	// Handle device fingerprint override
	if !plan.DevIdOverride.IsNull() && !plan.DevIdOverride.IsUnknown() {
		mac := plan.MAC.ValueString()
		device := int(plan.DevIdOverride.ValueInt64())

		err := c.OverrideUserFingerprint(ctx, site, mac, device)
		if err != nil {
			resp.Diagnostics.AddError("Error overriding user fingerprint", err.Error())
			return
		}

		result.DevIdOverride = device
	}

	resp.Diagnostics.Append(plan.Merge(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve meta-attributes from plan (they are not in the API response)
	plan.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.GetClient()
	site := c.ResolveSite(&state)
	id := state.GetID()

	result, err := c.GetUser(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	// The IP address is only available via GetUserByMAC, so issue a second request
	macResp, err := c.GetUserByMAC(ctx, site, result.MAC)
	if errors.Is(err, unifi.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading user by MAC", err.Error())
		return
	}

	result.IP = macResp.IP

	// Save meta-attributes before merge overwrites them
	allowExisting := state.AllowExisting
	skipForgetOnDestroy := state.SkipForgetOnDestroy

	resp.Diagnostics.Append(state.Merge(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore meta-attributes (not returned by API)
	state.AllowExisting = allowExisting
	state.SkipForgetOnDestroy = skipForgetOnDestroy
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.GetClient()
	site := c.ResolveSite(&plan)

	// Handle block/unblock changes
	planBlocked := plan.Blocked.ValueBool()
	stateBlocked := state.Blocked.ValueBool()
	if planBlocked != stateBlocked {
		mac := plan.MAC.ValueString()
		if planBlocked {
			err := c.BlockUserByMAC(ctx, site, mac)
			if err != nil {
				resp.Diagnostics.AddError("Error blocking user", err.Error())
				return
			}
		} else {
			err := c.UnblockUserByMAC(ctx, site, mac)
			if err != nil {
				resp.Diagnostics.AddError("Error unblocking user", err.Error())
				return
			}
		}
	}

	// Handle dev_id_override changes
	planDevId := plan.DevIdOverride.ValueInt64()
	stateDevId := state.DevIdOverride.ValueInt64()
	devIdChanged := plan.DevIdOverride.IsNull() != state.DevIdOverride.IsNull() || planDevId != stateDevId
	if devIdChanged {
		mac := plan.MAC.ValueString()
		device := int(planDevId)

		err := c.OverrideUserFingerprint(ctx, site, mac, device)
		if err != nil {
			resp.Diagnostics.AddError("Error overriding user fingerprint", err.Error())
			return
		}
	}

	body, diags := plan.AsUnifiModel(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	userReq := body.(*unifi.User)
	userReq.ID = state.GetID()
	userReq.SiteID = site

	result, err := c.UpdateUser(ctx, site, userReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	resp.Diagnostics.Append(plan.Merge(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipForgetOnDestroy.ValueBool() {
		return
	}

	c := r.GetClient()
	site := c.ResolveSite(&state)
	id := state.GetID()

	// Look up MAC instead of trusting state
	u, err := c.GetUser(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading user for delete", err.Error())
		return
	}

	err = c.DeleteUserByMAC(ctx, site, u.MAC)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.GetClient()
	if site == "" {
		site = c.Site
	}

	result, err := c.GetUser(ctx, site, id)
	if err != nil {
		resp.Diagnostics.AddError("Error importing user", err.Error())
		return
	}

	// The IP address is only available via GetUserByMAC
	macResp, err := c.GetUserByMAC(ctx, site, result.MAC)
	if err != nil {
		resp.Diagnostics.AddError("Error reading user by MAC during import", err.Error())
		return
	}
	result.IP = macResp.IP

	var state userModel
	resp.Diagnostics.Append(state.Merge(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults for meta-attributes on import
	state.AllowExisting = types.BoolValue(true)
	state.SkipForgetOnDestroy = types.BoolValue(false)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
