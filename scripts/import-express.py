#!/usr/bin/env python3
"""
UniFi Express State Import Tool

Queries a UniFi Express controller API to discover all existing resources
and generates Terraform import blocks + skeleton config.

Usage:
    export UNIFI_API=https://192.168.1.1:8443
    export UNIFI_USERNAME=admin
    export UNIFI_PASSWORD=yourpassword
    python3 scripts/import-express.py [--output-dir ./import]

Then:
    cd <output-dir>
    terraform plan -generate-config-out=generated.tf
"""

import argparse
import json
import os
import re
import ssl
import sys
import urllib.request
import urllib.error
import http.cookiejar


class UniFiClient:
    """Minimal UniFi API client for resource discovery."""

    def __init__(self, base_url, username, password, site="default", verify_ssl=False):
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.site = site

        ctx = ssl.create_default_context()
        if not verify_ssl:
            ctx.check_hostname = False
            ctx.verify_mode = ssl.CERT_NONE

        self.cookie_jar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPSHandler(context=ctx),
            urllib.request.HTTPCookieProcessor(self.cookie_jar),
        )

    def login(self):
        data = json.dumps({"username": self.username, "password": self.password}).encode()
        req = urllib.request.Request(
            f"{self.base_url}/api/login",
            data=data,
            headers={"Content-Type": "application/json"},
        )
        try:
            resp = self.opener.open(req)
            body = json.loads(resp.read())
            if body.get("meta", {}).get("rc") != "ok":
                print(f"Login failed: {body}", file=sys.stderr)
                sys.exit(1)
            print(f"Logged in to {self.base_url}")
        except urllib.error.URLError as e:
            print(f"Connection failed: {e}", file=sys.stderr)
            sys.exit(1)

    def get(self, path):
        url = f"{self.base_url}{path}"
        req = urllib.request.Request(url, headers={"Content-Type": "application/json"})
        try:
            resp = self.opener.open(req)
            body = json.loads(resp.read())
            return body.get("data", [])
        except urllib.error.HTTPError as e:
            if e.code == 404:
                return []
            print(f"  Warning: GET {path} returned {e.code}", file=sys.stderr)
            return []
        except Exception as e:
            print(f"  Warning: GET {path} failed: {e}", file=sys.stderr)
            return []

    def get_rest(self, endpoint):
        return self.get(f"/api/s/{self.site}/rest/{endpoint}")

    def get_stat(self, endpoint):
        return self.get(f"/api/s/{self.site}/stat/{endpoint}")

    def get_settings(self):
        return self.get(f"/api/s/{self.site}/get/setting")

    def get_sites(self):
        return self.get("/api/self/sites")


def sanitize_name(name):
    """Convert a resource name to a valid Terraform identifier."""
    if not name:
        return "unnamed"
    s = re.sub(r"[^a-zA-Z0-9_]", "_", name.lower())
    s = re.sub(r"_+", "_", s).strip("_")
    if s and s[0].isdigit():
        s = "r_" + s
    return s or "unnamed"


def deduplicate_names(items):
    """Ensure all (tf_type, tf_name) pairs are unique."""
    seen = {}
    for item in items:
        key = (item["tf_type"], item["tf_name"])
        if key in seen:
            seen[key] += 1
            item["tf_name"] = f"{item['tf_name']}_{seen[key]}"
        else:
            seen[key] = 0


# Each discoverer returns a list of dicts:
#   {"tf_type": "unifi_network", "tf_name": "my_lan", "import_id": "default:abc123"}

def discover_networks(client):
    resources = []
    for net in client.get_rest("networkconf"):
        name = net.get("name", "")
        _id = net.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_network",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_wlans(client):
    resources = []
    for wlan in client.get_rest("wlanconf"):
        name = wlan.get("name", "")
        _id = wlan.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_wlan",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_firewall_rules(client):
    resources = []
    for rule in client.get_rest("firewallrule"):
        name = rule.get("name", f"rule_{rule.get('rule_index', '')}")
        _id = rule.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_firewall_rule",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_firewall_groups(client):
    resources = []
    for group in client.get_rest("firewallgroup"):
        name = group.get("name", "")
        _id = group.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_firewall_group",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_firewall_zones(client):
    resources = []
    for zone in client.get_rest("firewallzone"):
        name = zone.get("name", "")
        _id = zone.get("_id", "")
        if not _id:
            continue
        # Framework resource: requires site:id format
        resources.append({
            "tf_type": "unifi_firewall_zone",
            "tf_name": sanitize_name(name),
            "import_id": f"{client.site}:{_id}",
            "display_name": name,
        })
    return resources


def discover_firewall_zone_policies(client):
    resources = []
    for policy in client.get_rest("firewallzonepolicy"):
        name = policy.get("name", "")
        _id = policy.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_firewall_zone_policy",
            "tf_name": sanitize_name(name or f"policy_{_id[:8]}"),
            "import_id": f"{client.site}:{_id}",
            "display_name": name,
        })
    return resources


def discover_port_forwards(client):
    resources = []
    for pf in client.get_rest("portforward"):
        name = pf.get("name", "")
        _id = pf.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_port_forward",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_static_routes(client):
    resources = []
    for route in client.get_rest("routing"):
        name = route.get("name", "")
        _id = route.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_static_route",
            "tf_name": sanitize_name(name or f"route_{_id[:8]}"),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_port_profiles(client):
    resources = []
    for profile in client.get_rest("portconf"):
        name = profile.get("name", "")
        _id = profile.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_port_profile",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_devices(client):
    resources = []
    for device in client.get_stat("device"):
        name = device.get("name", device.get("model", ""))
        mac = device.get("mac", "")
        _id = device.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_device",
            "tf_name": sanitize_name(name or f"device_{_id[:8]}"),
            "import_id": mac if mac else _id,
            "display_name": f"{name} ({mac})",
        })
    return resources


def discover_user_groups(client):
    resources = []
    for group in client.get_rest("usergroup"):
        name = group.get("name", "")
        _id = group.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_user_group",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_users(client):
    resources = []
    for user in client.get_rest("user"):
        name = user.get("name", user.get("mac", ""))
        _id = user.get("_id", "")
        if not _id:
            continue
        # Skip users without a fixed IP or name (transient clients)
        if not user.get("name") and not user.get("fixed_ip"):
            continue
        resources.append({
            "tf_type": "unifi_user",
            "tf_name": sanitize_name(name or f"user_{_id[:8]}"),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_dns_records(client):
    resources = []
    for record in client.get_rest("dnsrecord"):
        name = record.get("key", record.get("value", ""))
        _id = record.get("_id", "")
        if not _id:
            continue
        # Framework resource: requires site:id
        resources.append({
            "tf_type": "unifi_dns_record",
            "tf_name": sanitize_name(name),
            "import_id": f"{client.site}:{_id}",
            "display_name": name,
        })
    return resources


def discover_dynamic_dns(client):
    resources = []
    for ddns in client.get_rest("dynamicdns"):
        name = ddns.get("host_name", "")
        _id = ddns.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_dynamic_dns",
            "tf_name": sanitize_name(name or f"ddns_{_id[:8]}"),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_accounts(client):
    resources = []
    for acct in client.get_rest("account"):
        name = acct.get("name", "")
        _id = acct.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_account",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


def discover_radius_profiles(client):
    resources = []
    for profile in client.get_rest("radiusprofile"):
        name = profile.get("name", "")
        _id = profile.get("_id", "")
        if not _id:
            continue
        resources.append({
            "tf_type": "unifi_radius_profile",
            "tf_name": sanitize_name(name),
            "import_id": _id,
            "display_name": name,
        })
    return resources


# Settings are singletons per site. The import ID is site:settings_id.
SETTING_KEY_TO_TF_TYPE = {
    "auto_speedtest": "unifi_setting_auto_speedtest",
    "country":        "unifi_setting_country",
    "dpi":            "unifi_setting_dpi",
    "guest_access":   "unifi_setting_guest_access",
    "locale":         "unifi_setting_locale",
    "super_mail":     None,  # not a terraform resource
    "ntp":            "unifi_setting_ntp",
    "rsyslogd":       "unifi_setting_rsyslogd",
    "teleport":       "unifi_setting_teleport",
    "usg":            "unifi_setting_usg",
    "usw":            "unifi_setting_usw",
    "network_optimization": "unifi_setting_network_optimization",
    "magic_site_to_site_vpn": "unifi_setting_magic_site_to_site_vpn",
    "radius":         "unifi_setting_radius",
    "mgmt":           None,  # mgmt does not support import
    "connectivity":   None,  # not a terraform resource
    "super_events":   None,
    "super_identity": None,
    "super_sdn":      None,
    "super_smtp":     None,
    "snmp":           None,
    "element_adopt":  None,
    "porta":          None,
    "global_ap":      None,
    "global_switch":  None,
}


def discover_settings(client):
    resources = []
    settings = client.get_settings()
    for setting in settings:
        key = setting.get("key", "")
        _id = setting.get("_id", "")
        tf_type = SETTING_KEY_TO_TF_TYPE.get(key)
        if not tf_type or not _id:
            continue
        # radius uses SDK v2 (bare id works), all others use Framework (site:id required)
        if tf_type == "unifi_setting_radius":
            import_id = _id
        else:
            import_id = f"{client.site}:{_id}"
        resources.append({
            "tf_type": tf_type,
            "tf_name": sanitize_name(key),
            "import_id": import_id,
            "display_name": f"Setting: {key}",
        })
    return resources


def generate_imports_tf(resources, output_dir):
    """Generate imports.tf with Terraform import blocks."""
    lines = [
        "# Auto-generated import blocks from UniFi Express",
        "# Run: terraform plan -generate-config-out=generated.tf",
        "",
    ]
    for r in resources:
        lines.append(f"# {r['display_name']}")
        lines.append(f"import {{")
        lines.append(f'  to = {r["tf_type"]}.{r["tf_name"]}')
        lines.append(f'  id = "{r["import_id"]}"')
        lines.append(f"}}")
        lines.append("")

    path = os.path.join(output_dir, "imports.tf")
    with open(path, "w") as f:
        f.write("\n".join(lines))
    return path


def generate_provider_tf(output_dir):
    """Generate provider.tf with the provider configuration."""
    content = '''\
terraform {
  required_providers {
    unifi = {
      source = "p3terp4N/unifi-express"
    }
  }
}

provider "unifi" {
  username       = var.unifi_username
  password       = var.unifi_password
  api_url        = var.unifi_api_url
  allow_insecure = true
  site           = var.unifi_site
}

variable "unifi_username" {
  type    = string
  default = ""
}

variable "unifi_password" {
  type      = string
  sensitive = true
  default   = ""
}

variable "unifi_api_url" {
  type    = string
  default = ""
}

variable "unifi_site" {
  type    = string
  default = "default"
}
'''
    path = os.path.join(output_dir, "provider.tf")
    with open(path, "w") as f:
        f.write(content)
    return path


def main():
    parser = argparse.ArgumentParser(
        description="Discover UniFi Express resources and generate Terraform import blocks"
    )
    parser.add_argument(
        "--output-dir", "-o",
        default="./import",
        help="Directory to write generated .tf files (default: ./import)",
    )
    parser.add_argument(
        "--site", "-s",
        default=os.environ.get("UNIFI_SITE", "default"),
        help="UniFi site name (default: $UNIFI_SITE or 'default')",
    )
    parser.add_argument(
        "--skip-settings",
        action="store_true",
        help="Skip importing settings resources",
    )
    parser.add_argument(
        "--skip-users",
        action="store_true",
        help="Skip importing user resources",
    )
    args = parser.parse_args()

    api_url = os.environ.get("UNIFI_API", "")
    username = os.environ.get("UNIFI_USERNAME", "")
    password = os.environ.get("UNIFI_PASSWORD", "")

    if not api_url or not username or not password:
        print("Error: Set UNIFI_API, UNIFI_USERNAME, and UNIFI_PASSWORD environment variables.", file=sys.stderr)
        print("Example:", file=sys.stderr)
        print("  export UNIFI_API=https://192.168.1.1:8443", file=sys.stderr)
        print("  export UNIFI_USERNAME=admin", file=sys.stderr)
        print("  export UNIFI_PASSWORD=yourpassword", file=sys.stderr)
        sys.exit(1)

    client = UniFiClient(api_url, username, password, site=args.site)
    client.login()

    # Discover all resources
    all_resources = []

    discoverers = [
        ("Networks",              discover_networks),
        ("WLANs",                 discover_wlans),
        ("Firewall Rules",        discover_firewall_rules),
        ("Firewall Groups",       discover_firewall_groups),
        ("Firewall Zones",        discover_firewall_zones),
        ("Firewall Zone Policies", discover_firewall_zone_policies),
        ("Port Forwards",         discover_port_forwards),
        ("Static Routes",         discover_static_routes),
        ("Port Profiles",         discover_port_profiles),
        ("Devices",               discover_devices),
        ("User Groups",           discover_user_groups),
        ("DNS Records",           discover_dns_records),
        ("Dynamic DNS",           discover_dynamic_dns),
        ("RADIUS Accounts",       discover_accounts),
        ("RADIUS Profiles",       discover_radius_profiles),
    ]

    if not args.skip_users:
        discoverers.append(("Users", discover_users))

    if not args.skip_settings:
        discoverers.append(("Settings", discover_settings))

    for label, discoverer in discoverers:
        print(f"Discovering {label}...", end=" ")
        resources = discoverer(client)
        print(f"found {len(resources)}")
        all_resources.extend(resources)

    if not all_resources:
        print("\nNo resources found. Nothing to import.")
        sys.exit(0)

    # Deduplicate terraform names
    deduplicate_names(all_resources)

    # Generate output
    os.makedirs(args.output_dir, exist_ok=True)

    imports_path = generate_imports_tf(all_resources, args.output_dir)
    provider_path = generate_provider_tf(args.output_dir)

    print(f"\n{'='*60}")
    print(f"Discovered {len(all_resources)} resources")
    print(f"Generated: {imports_path}")
    print(f"Generated: {provider_path}")
    print(f"{'='*60}")
    print()
    print("Next steps:")
    print(f"  cd {args.output_dir}")
    print()
    print("  # Set your credentials:")
    print(f"  export TF_VAR_unifi_username=$UNIFI_USERNAME")
    print(f"  export TF_VAR_unifi_password=$UNIFI_PASSWORD")
    print(f"  export TF_VAR_unifi_api_url=$UNIFI_API")
    print()
    print("  # Generate full Terraform config from imported state:")
    print("  terraform plan -generate-config-out=generated.tf")
    print()
    print("  # Review generated.tf, then:")
    print("  terraform apply")
    print()
    print("Note: Some generated config may need manual tweaks.")
    print("Run 'terraform plan' after to verify zero diff.")


if __name__ == "__main__":
    main()
