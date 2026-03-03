terraform {
  required_providers {
    unifi = {
      source = "p3terp4N/unifi-express"
    }
  }
}

# All values from env vars: UNIFI_USERNAME, UNIFI_PASSWORD, UNIFI_API, UNIFI_INSECURE
provider "unifi" {}
