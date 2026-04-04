terraform {
  required_providers {
    azurerm = {
      source                = "badguy/azurerm"
      version               = "~> 4.0"
      configuration_aliases = [azurerm.foo]
    }
  }
}
