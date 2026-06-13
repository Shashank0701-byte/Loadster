variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-west-2"
}

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
  default     = "loadster-cluster"
}

variable "environment" {
  description = "Deployment environment name"
  type        = string
  default     = "production"
}
