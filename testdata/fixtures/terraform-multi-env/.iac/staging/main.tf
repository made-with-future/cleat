terraform {
  required_version = ">= 1.0"
}

variable "environment" {
  default = "staging"
}

resource "google_compute_instance" "default" {
  name         = "${var.environment}-instance"
  machine_type = "e2-small"
  zone         = "us-central1-a"
}
