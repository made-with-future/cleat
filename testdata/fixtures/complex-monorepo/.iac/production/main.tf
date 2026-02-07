terraform {
  required_version = ">= 1.0"
  backend "gcs" {
    bucket = "complex-project-terraform"
    prefix = "production"
  }
}

provider "google" {
  project = "complex-project"
  region  = "us-central1"
}

resource "google_compute_instance" "app" {
  name         = "production-app"
  machine_type = "e2-medium"
  zone         = "us-central1-a"
}
