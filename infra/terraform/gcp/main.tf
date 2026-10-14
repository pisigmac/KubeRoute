terraform {
  required_providers {
    google = { source = "hashicorp/google", version = "~> 5.0" }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" { default = "kuberoute-prod" }
variable "region" { default = "us-central1" }

resource "google_container_cluster" "control_plane" {
  name     = "kuberoute-control-plane"
  location = var.region
  initial_node_count = 3
}
