terraform {
  required_version = ">= 1.0"

  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "production"
  }
}

provider "google" {
  project = "my-project"
  region  = "us-central1"
}

resource "google_compute_instance" "default" {
  name         = "production-instance"
  machine_type = "e2-medium"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  network_interface {
    network = "default"
  }
}
