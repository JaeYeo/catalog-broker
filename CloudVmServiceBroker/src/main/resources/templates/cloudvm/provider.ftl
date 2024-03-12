# Define required providers
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.42.0"
    }
  }
}

# Configure the OpenStack Provider
provider "openstack" {
  user_name   = "${user_name}"
  tenant_name = "${tenant_name}"
  domain_name = "${domain_name}"
  password    = "${password}"
  auth_url    = "${auth_url}"
  region      = "${region}"
}