terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
	region = "eu-west-2"
}

provider "aws" {
	region = "eu-west-1"
	alias = "ie"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

variable "CLIENT_ID" {
  type = string 
  nullable = false
}

variable "CLIENT_SECRET" {
  type     = string
  nullable = false
}

variable "REDIRECT_URI" {
  type     = string
  nullable = false
}
variable "MAIL_DOMAIN" {
  type     = string
  nullable = false
}
variable "MAIL_RECIPIENT" {
  type     = string
  nullable = false
}
variable "SMTP_HOST" {
  type     = string
  nullable = false
}
variable "SMTP_PORT" {
  type     = string
  nullable = false
}
variable "SMTP_USER" {
  type     = string
  nullable = false
}
variable "SMTP_PASSWORD" {
  type     = string
  nullable = false
}