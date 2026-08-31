variable "cluster_name" {
  description = "Name of the Kind cluster"
  type        = string
  default     = "realestate-trust"
}

variable "project_root" {
  description = "Absolute path to the project root for building images"
  type        = string
  default     = "../.."
}
