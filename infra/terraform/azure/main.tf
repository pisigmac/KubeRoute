terraform {
  required_providers {
    azurerm = { source = "hashicorp/azurerm", version = "~> 3.0" }
  }
}

provider "azurerm" {
  features {}
}

variable "region" { default = "East US" }

resource "azurerm_kubernetes_cluster" "control_plane" {
  name                = "kuberoute-control-plane"
  location            = var.region
  resource_group_name = "kuberoute-rg"
  dns_prefix          = "kuberoute"

  default_node_pool {
    name       = "default"
    node_count = 3
    vm_size    = "Standard_D2_v2"
  }

  identity {
    type = "SystemAssigned"
  }
}
