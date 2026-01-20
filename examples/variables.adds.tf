variable "configure_adds_resources" {
  description = <<DESCRIPTION
Configures the adds resources for the AzERE deployment.

### Attributes

<!-- MARINATED: configure\_adds\_resources -->

- `advanced` - (Optional) 
- `location` - (Optional) 
- `settings` - (Optional) # TODO: Add description for settings
  - `forests` - (Optional) # TODO: Add description for forests
    - `config` - (Optional) # TODO: Add description for config
      - `backup_policy` - (Optional) # TODO: Add description for backup_policy
        - `enhanced` - (Optional) # TODO: Add description for enhanced
      - `network` - (Optional) # TODO: Add description for network
      - `network_security_rules` - (Optional) # TODO: Add description for network_security_rules
        - `custom_prefixes` - (Optional) # TODO: Add description for custom_prefixes
        - `remote_address_prefixes` - (Optional) # TODO: Add description for remote_address_prefixes
      - `recovery_service` - (Optional) # TODO: Add description for recovery_service
        - `immutability` - (Optional) # TODO: Add description for immutability
    - `domains` - (Optional) # TODO: Add description for domains
      - `config` - (Optional) # TODO: Add description for config
        - `domain_controllers` - (Optional) # TODO: Add description for domain_controllers
          - `azure_update` - (Optional) # TODO: Add description for azure_update
            - `enabled` - (Optional) # TODO: Add description for enabled
            - `template` - (Optional) # TODO: Add description for template
          - `computer_name` - (Optional) # TODO: Add description for computer_name
          - `dns_servers` - (Optional) # TODO: Add description for dns_servers
          - `enable_boot_diagnostics` - (Optional) # TODO: Add description for enable_boot_diagnostics
          - `enabled` - (Optional) # TODO: Add description for enabled
          - `image_sku` - (Optional) # TODO: Add description for image_sku
          - `meta_data` - (Optional) # TODO: Add description for meta_data
            - `tags` - (Optional) # TODO: Add description for tags
          - `name` - (Optional) # TODO: Add description for name
          - `role_assignments` - (Optional) # TODO: Add description for role_assignments
            - `principal_id` - (Required) # TODO: Add description for principal_id
            - `role_definition_name` - (Required) # TODO: Add description for role_definition_name
          - `size` - (Optional) # TODO: Add description for size
          - `use_enhanced_backup` - (Optional) # TODO: Add description for use_enhanced_backup
          - `use_trusted_launch` - (Optional) # TODO: Add description for use_trusted_launch
        - `domain_type` - (Optional) # TODO: Add description for domain_type
        - `network` - (Optional) # TODO: Add description for network
        - `network_security_rules` - (Optional) # TODO: Add description for network_security_rules
          - `custom_prefixes` - (Optional) # TODO: Add description for custom_prefixes
          - `remote_address_prefixes` - (Optional) # TODO: Add description for remote_address_prefixes
        - `optional_routes` - (Optional) # TODO: Add description for optional_routes
        - `parent_domain_name` - (Optional) # TODO: Add description for parent_domain_name
        - `users` - (Optional) # TODO: Add description for users
          - `customer_role_assignment` - (Optional) # TODO: Add description for customer_role_assignment
      - `fqdn` - (Optional) # TODO: Add description for fqdn
      - `index` - (Optional) # TODO: Add description for index
      - `netbios_name` - (Optional) # TODO: Add description for netbios_name
    - `enabled` - (Optional) # TODO: Add description for enabled
    - `fqdn` - (Optional) # TODO: Add description for fqdn
    - `index` - (Optional) # TODO: Add description for index
    - `netbios_name` - (Optional) # TODO: Add description for netbios_name
    - `test` - (Optional) # TODO: Add description for test
  - `maintenance_templates` - (Optional) # TODO: Add description for maintenance_templates
    - `classifications_to_include` - (Optional) # TODO: Add description for classifications_to_include
    - `duration` - (Optional) # TODO: Add description for duration
    - `expiration_date_time` - (Optional) # TODO: Add description for expiration_date_time
    - `in_guest_user_patch_mode` - (Optional) # TODO: Add description for in_guest_user_patch_mode
    - `name` - (Required) # TODO: Add description for name
    - `reboot` - (Optional) # TODO: Add description for reboot
    - `recur_every` - (Required) # TODO: Add description for recur_every
    - `scope` - (Optional) # TODO: Add description for scope
    - `start_date_time` - (Required) # TODO: Add description for start_date_time
    - `time_zone` - (Optional) # TODO: Add description for time_zone
    - `visibility` - (Optional) # TODO: Add description for visibility
- `tags` - (Optional) 


<!-- /MARINATED: configure\_adds\_resources -->
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
              enabled       = optional(bool, true)
              name          = optional(string, "")
              computer_name = optional(string, "")
              size          = optional(string, "")
              image_sku     = optional(string, "")
              data_disk_size = optional(number, 20)
              dns_servers = optional(list(string), [])
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
