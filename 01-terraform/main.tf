terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=3.0.0"
    }
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
}

terraform {
  backend "azurerm" {
    storage_account_name = "asdf"
    container_name       = "tfstate"
    key                  = "01-terraform.tfstate"
    use_azuread_auth     = true
  }
}

resource "azurerm_resource_group" "rg-01-terraform" {
  name     = "rg-tf-dagger"
  location = "West Europe"
}