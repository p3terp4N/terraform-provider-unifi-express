# UniFi Express Terraform Provider

A Terraform provider for managing Ubiquiti's **UniFi Express** devices running **Network Application 8.x**. This is a fork of [filipowm/terraform-provider-unifi](https://github.com/filipowm/terraform-provider-unifi), stripped down to only the resources supported by the UniFi Express hardware.

**Note:** You can't configure your network while connected to something that may disconnect (like WiFi).
Use a hard-wired connection to your controller to use this provider.

## Features

- Manage UniFi Express network resources using Infrastructure as Code
- Targets **Network Application 8.x** (the version running on UniFi Express)
- Username/password authentication only (API keys require 9.0.108+, which is not available on Express)
- 30+ resources including networks, WLANs, firewall rules/groups/zones, port profiles, DNS records, static routes, and settings
- Removed unsupported resources: IDS/IPS, SSL Inspection, LCM (no LCD panel on Express)

## Installation (Local Dev)

This provider is not published to the Terraform Registry. Use `dev_overrides` for local development:

```bash
# Build and install
cd terraform-provider-unifi-express
go install .
```

Add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "p3terp4N/unifi-express" = "/Users/<YOU>/go/bin"
  }
  direct {}
}
```

## Usage

```hcl
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
  api_url        = "https://192.168.1.1:8443"
  allow_insecure = true
  site           = "default"
}
```

## Authentication

The provider supports **username/password authentication only**:

```bash
export UNIFI_USERNAME="admin"
export UNIFI_PASSWORD="password"
export UNIFI_API="https://192.168.1.1:8443"
export UNIFI_INSECURE=true
```

Or configure directly in the provider block:

```hcl
provider "unifi" {
  username       = "admin"
  password       = "password"
  api_url        = "https://192.168.1.1:8443"
  allow_insecure = true
  site           = "default"
}
```

## Example

```hcl
resource "unifi_network" "vlan_50" {
  name    = "VLAN 50"
  purpose = "corporate"
  subnet  = "10.0.50.0/24"
  vlan_id = 50
}

resource "unifi_wlan" "wifi" {
  name       = "My WiFi Network"
  security   = "wpapsk"
  passphrase = "mystrongpassword"
  network_id = unifi_network.vlan_50.id
}
```

## Supported Platform

- **UniFi Express** running Network Application 8.x

The provider enforces version constraints and will reject connections to controllers outside the 8.x range.

## Testing

Against a real UniFi Express device:

```bash
export UNIFI_API=https://<express-ip>:8443
export UNIFI_USERNAME=<admin>
export UNIFI_PASSWORD=<pass>
export UNIFI_INSECURE=true
TF_ACC=1 go test ./internal/provider/acctest/... -v -timeout 30m
```

Against Docker (for development without hardware):

```bash
docker compose up -d
TF_ACC=1 go test ./internal/provider/acctest/... -v -timeout 30m
```

## Acknowledgements

This project is a fork of [filipowm/terraform-provider-unifi](https://github.com/filipowm/terraform-provider-unifi), which itself is a fork of [paultyng/terraform-provider-unifi](https://github.com/paultyng/terraform-provider-unifi). We extend our gratitude to both maintainers and all contributors for their foundational work.

The provider is built on top of the [go-unifi](https://github.com/filipowm/go-unifi) SDK.

## License

This provider is licensed under the [LICENSE](./LICENSE) file.
