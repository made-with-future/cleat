terraform {
  required_version = ">= 1.0"
}

resource "google_compute_instance" "app" {
  name         = "staging-app"
  machine_type = "e2-small"
  zone         = "us-central1-a"
}
