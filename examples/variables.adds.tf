variable "configure_adds_resources" {
  description = <<DESCRIPTION
Configures the adds resources for the AzERE deployment.

### Attributes

<!-- MARINATED: configure_adds_resources -->
<!-- /MARINATED: configure_adds_resources -->

DESCRIPTION
  type = object({
    settings = optional(object({
      maintenance_templates = optional(list(object({
        name                       = string
        scope                      = optional(string, "InGuestPatch")
        in_guest_user_patch_mode   = optional(string, "User")
        visibility                 = optional(string, "Custom")
        start_date_time            = string
        expiration_date_time       = optional(string)
        duration                   = optional(string, "02:00")
        time_zone                  = optional(string, "UTC")
        recur_every                = string
        reboot                     = optional(string, "IfRequired")
        classifications_to_include = optional(list(string), ["Critical", "Security"])
      })), [])
      forests = optional(list(object({
        enabled      = optional(bool, true)
        test         = optional(string, "")
        netbios_name = optional(string, "")
        fqdn         = optional(string, "")
        index        = optional(number, 1)
        config = optional(object({
          recovery_service = optional(object({
            immutability = optional(string)
          }), {})
          backup_policy = optional(object({
            enhanced = optional(bool, false)
          }), {})
          network = optional(object({
          }), {})
          network_security_rules = optional(object({
            remote_address_prefixes = optional(list(string))
            custom_prefixes         = optional(any)
          }), {})
        }), {})
        domains = optional(list(object({
          netbios_name = optional(string, "")
          fqdn         = optional(string, "")
          index        = optional(number, 1)
          config = optional(object({
            domain_type        = optional(string, "")
            parent_domain_name = optional(string, "")
            network = optional(object({
            }), {})
            network_security_rules = optional(object({
              remote_address_prefixes = optional(list(string))
              custom_prefixes         = optional(any)
            }), {})
            users = optional(object({
              customer_role_assignment = optional(list(string), [])
            }), {})
            optional_routes = optional(list(string), [])
            domain_controllers = optional(list(object({
              enabled        = optional(bool, true)
              name           = optional(string, "")
              computer_name  = optional(string, "")
              size           = optional(string, "")
              image_sku      = optional(string, "")
              data_disk_size = optional(number, 20)
              dns_servers    = optional(list(string), [])
              meta_data = optional(object({
                tags = optional(map(string))
              }), {})
              use_enhanced_backup     = optional(bool, false)
              use_trusted_launch      = optional(bool, false)
              enable_boot_diagnostics = optional(bool, false)
              azure_update = optional(object({
                enabled  = optional(bool, false)
                template = optional(string, "")
              }), {})
              role_assignments = optional(list(object({
                role_definition_name = string
                principal_id         = string
              })), []),
            })))
          }), {})
        })), [])
      })), [])
    }), {})
    location = optional(string, "")
    tags     = optional(any, {})
    advanced = optional(any, {})
  })
  default = {}
}
