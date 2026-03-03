package network

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
)

var _ base.ResourceModel = &wlanModel{}

type wlanModel struct {
	base.Model
	Name                   types.String `tfsdk:"name"`
	UserGroupID            types.String `tfsdk:"user_group_id"`
	Security               types.String `tfsdk:"security"`
	WPA3Support            types.Bool   `tfsdk:"wpa3_support"`
	WPA3Transition         types.Bool   `tfsdk:"wpa3_transition"`
	PMFMode                types.String `tfsdk:"pmf_mode"`
	Passphrase             types.String `tfsdk:"passphrase"`
	HideSSID               types.Bool   `tfsdk:"hide_ssid"`
	IsGuest                types.Bool   `tfsdk:"is_guest"`
	MulticastEnhance       types.Bool   `tfsdk:"multicast_enhance"`
	MACFilterEnabled       types.Bool   `tfsdk:"mac_filter_enabled"`
	MACFilterList          types.Set    `tfsdk:"mac_filter_list"`
	MACFilterPolicy        types.String `tfsdk:"mac_filter_policy"`
	RadiusProfileID        types.String `tfsdk:"radius_profile_id"`
	Schedule               types.List   `tfsdk:"schedule"`
	No2GhzOui              types.Bool   `tfsdk:"no2ghz_oui"`
	L2Isolation            types.Bool   `tfsdk:"l2_isolation"`
	ProxyArp               types.Bool   `tfsdk:"proxy_arp"`
	BssTransition          types.Bool   `tfsdk:"bss_transition"`
	Uapsd                  types.Bool   `tfsdk:"uapsd"`
	FastRoamingEnabled     types.Bool   `tfsdk:"fast_roaming_enabled"`
	MinimumDataRate2gKbps  types.Int64  `tfsdk:"minimum_data_rate_2g_kbps"`
	MinimumDataRate5gKbps  types.Int64  `tfsdk:"minimum_data_rate_5g_kbps"`
	WLANBand               types.String `tfsdk:"wlan_band"`
	NetworkID              types.String `tfsdk:"network_id"`
	ApGroupIDs             types.Set    `tfsdk:"ap_group_ids"`
}

type scheduleModel struct {
	DayOfWeek   types.String `tfsdk:"day_of_week"`
	StartHour   types.Int64  `tfsdk:"start_hour"`
	StartMinute types.Int64  `tfsdk:"start_minute"`
	Duration    types.Int64  `tfsdk:"duration"`
	Name        types.String `tfsdk:"name"`
}

func scheduleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"day_of_week":  types.StringType,
		"start_hour":   types.Int64Type,
		"start_minute": types.Int64Type,
		"duration":     types.Int64Type,
		"name":         types.StringType,
	}
}

func scheduleObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: scheduleAttrTypes()}
}

func (m *wlanModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	security := m.Security.ValueString()
	passphrase := m.Passphrase.ValueString()
	if security == "open" {
		passphrase = ""
	}

	wpa3 := m.WPA3Support.ValueBool()
	wpa3Transition := m.WPA3Transition.ValueBool()
	pmf := m.PMFMode.ValueString()

	if security != "wpapsk" {
		if wpa3 || wpa3Transition {
			diags.AddError("Invalid WPA3 configuration", "wpa3_support and wpa3_transition are only valid for security type wpapsk")
			return nil, diags
		}
	}

	if wpa3Transition && pmf == "disabled" {
		diags.AddError("Invalid PMF configuration", "WPA3 transition mode requires pmf_mode to be turned on.")
		return nil, diags
	} else if wpa3 && !wpa3Transition && pmf != "required" {
		diags.AddError("Invalid PMF configuration", "For WPA3 you must set pmf_mode to required.")
		return nil, diags
	}

	macFilterEnabled := m.MACFilterEnabled.ValueBool()
	var macFilterList []string
	if macFilterEnabled && !m.MACFilterList.IsNull() && !m.MACFilterList.IsUnknown() {
		diags.Append(m.MACFilterList.ElementsAs(ctx, &macFilterList, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	var apGroupIDs []string
	if !m.ApGroupIDs.IsNull() && !m.ApGroupIDs.IsUnknown() {
		diags.Append(m.ApGroupIDs.ElementsAs(ctx, &apGroupIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Process schedule
	var schedules []unifi.WLANScheduleWithDuration
	if !m.Schedule.IsNull() && !m.Schedule.IsUnknown() {
		var scheduleModels []scheduleModel
		diags.Append(m.Schedule.ElementsAs(ctx, &scheduleModels, false)...)
		if diags.HasError() {
			return nil, diags
		}
		for _, s := range scheduleModels {
			schedules = append(schedules, unifi.WLANScheduleWithDuration{
				StartDaysOfWeek: []string{s.DayOfWeek.ValueString()},
				StartHour:       int(s.StartHour.ValueInt64()),
				StartMinute:     int(s.StartMinute.ValueInt64()),
				DurationMinutes: int(s.Duration.ValueInt64()),
				Name:            s.Name.ValueString(),
			})
		}
	}

	rate2g := int(m.MinimumDataRate2gKbps.ValueInt64())
	rate5g := int(m.MinimumDataRate5gKbps.ValueInt64())

	minrateSettingPreference := "auto"
	if rate2g != 0 || rate5g != 0 {
		if rate2g == 0 || rate5g == 0 {
			diags.AddError("Invalid minimum data rate configuration", "you must set minimum data rates on both 2g and 5g if setting either")
			return nil, diags
		}
		minrateSettingPreference = "manual"
	}

	return &unifi.WLAN{
		ID:                      m.ID.ValueString(),
		Name:                    m.Name.ValueString(),
		XPassphrase:             passphrase,
		HideSSID:                m.HideSSID.ValueBool(),
		IsGuest:                 m.IsGuest.ValueBool(),
		NetworkID:               m.NetworkID.ValueString(),
		ApGroupIDs:              apGroupIDs,
		UserGroupID:             m.UserGroupID.ValueString(),
		Security:                security,
		WPA3Support:             wpa3,
		WPA3Transition:          wpa3Transition,
		MulticastEnhanceEnabled: m.MulticastEnhance.ValueBool(),
		MACFilterEnabled:        macFilterEnabled,
		MACFilterList:           macFilterList,
		MACFilterPolicy:         m.MACFilterPolicy.ValueString(),
		RADIUSProfileID:         m.RadiusProfileID.ValueString(),
		ScheduleWithDuration:    schedules,
		ScheduleEnabled:         len(schedules) > 0,
		WLANBand:                m.WLANBand.ValueString(),
		PMFMode:                 pmf,

		// Hardcoded defaults (same as V1)
		WPAEnc:             "ccmp",
		WPAMode:            "wpa2",
		Enabled:            true,
		NameCombineEnabled: true,
		GroupRekey:          3600,
		DTIMMode:           "default",

		No2GhzOui:          m.No2GhzOui.ValueBool(),
		L2Isolation:        m.L2Isolation.ValueBool(),
		ProxyArp:           m.ProxyArp.ValueBool(),
		BssTransition:      m.BssTransition.ValueBool(),
		UapsdEnabled:       m.Uapsd.ValueBool(),
		FastRoamingEnabled: m.FastRoamingEnabled.ValueBool(),

		MinrateSettingPreference: minrateSettingPreference,
		MinrateNgEnabled:         rate2g != 0,
		MinrateNgDataRateKbps:    rate2g,
		MinrateNaEnabled:         rate5g != 0,
		MinrateNaDataRateKbps:    rate5g,
	}, diags
}

func (m *wlanModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.WLAN)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.WLAN, got: %T", i))
		return diags
	}

	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.UserGroupID = types.StringValue(other.UserGroupID)

	security := other.Security
	passphrase := other.XPassphrase
	wpa3 := false
	wpa3Transition := false
	switch security {
	case "open":
		passphrase = ""
	case "wpapsk":
		wpa3 = other.WPA3Support
		wpa3Transition = other.WPA3Transition
	}

	m.Security = types.StringValue(security)
	if passphrase != "" {
		m.Passphrase = types.StringValue(passphrase)
	} else {
		m.Passphrase = types.StringNull()
	}
	m.WPA3Support = types.BoolValue(wpa3)
	m.WPA3Transition = types.BoolValue(wpa3Transition)
	m.PMFMode = types.StringValue(other.PMFMode)
	m.HideSSID = types.BoolValue(other.HideSSID)
	m.IsGuest = types.BoolValue(other.IsGuest)
	m.MulticastEnhance = types.BoolValue(other.MulticastEnhanceEnabled)

	macFilterEnabled := other.MACFilterEnabled
	m.MACFilterEnabled = types.BoolValue(macFilterEnabled)
	if macFilterEnabled {
		macSet, d := types.SetValueFrom(ctx, types.StringType, other.MACFilterList)
		diags.Append(d...)
		if !diags.HasError() {
			m.MACFilterList = macSet
		}
		m.MACFilterPolicy = types.StringValue(other.MACFilterPolicy)
	} else {
		m.MACFilterList = types.SetNull(types.StringType)
		m.MACFilterPolicy = types.StringValue("deny")
	}

	m.RadiusProfileID = ut.StringOrNull(other.RADIUSProfileID)

	// Convert schedule from API response
	if len(other.ScheduleWithDuration) > 0 {
		var scheduleModels []scheduleModel
		for _, s := range other.ScheduleWithDuration {
			for _, dow := range s.StartDaysOfWeek {
				scheduleModels = append(scheduleModels, scheduleModel{
					DayOfWeek:   types.StringValue(dow),
					StartHour:   types.Int64Value(int64(s.StartHour)),
					StartMinute: types.Int64Value(int64(s.StartMinute)),
					Duration:    types.Int64Value(int64(s.DurationMinutes)),
					Name:        ut.StringOrNull(s.Name),
				})
			}
		}
		schedList, d := types.ListValueFrom(ctx, scheduleObjectType(), scheduleModels)
		diags.Append(d...)
		if !diags.HasError() {
			m.Schedule = schedList
		}
	} else {
		m.Schedule = types.ListNull(scheduleObjectType())
	}

	m.No2GhzOui = types.BoolValue(other.No2GhzOui)
	m.L2Isolation = types.BoolValue(other.L2Isolation)
	m.ProxyArp = types.BoolValue(other.ProxyArp)
	m.BssTransition = types.BoolValue(other.BssTransition)
	m.Uapsd = types.BoolValue(other.UapsdEnabled)
	m.FastRoamingEnabled = types.BoolValue(other.FastRoamingEnabled)
	m.WLANBand = types.StringValue(other.WLANBand)
	m.NetworkID = ut.StringOrNull(other.NetworkID)

	// AP Group IDs
	if len(other.ApGroupIDs) > 0 {
		apSet, d := types.SetValueFrom(ctx, types.StringType, other.ApGroupIDs)
		diags.Append(d...)
		if !diags.HasError() {
			m.ApGroupIDs = apSet
		}
	} else {
		m.ApGroupIDs = types.SetNull(types.StringType)
	}

	// Minimum data rates
	if other.MinrateSettingPreference != "auto" && other.MinrateNgEnabled {
		m.MinimumDataRate2gKbps = types.Int64Value(int64(other.MinrateNgDataRateKbps))
	} else {
		m.MinimumDataRate2gKbps = types.Int64Value(0)
	}
	if other.MinrateSettingPreference != "auto" && other.MinrateNaEnabled {
		m.MinimumDataRate5gKbps = types.Int64Value(int64(other.MinrateNaDataRateKbps))
	} else {
		m.MinimumDataRate5gKbps = types.Int64Value(0)
	}

	return diags
}
