# Terraform Configuration Documentation

### configure\_adds\_resources

Description: Configures the adds resources for the AzERE deployment.

### Attributes

<!-- MARINATED: configure\_adds\_resources -->

- **advanced** - (Optional)  A map of advanced settings to apply to underlying resources. See the advanced section for more details.
- **location** - (Optional) The Azure region in which region bound resources will be deployed.
- **settings** - (Optional) Settings for the Azure Active Directory Domain Services deployment
  - **forests** - (Optional) Configuration settings for Azure Active Directory Domain Services forests.
    - **config** - (Optional) Configuration options object for forest wide settings
      - **backup_policy** - (Optional) Configuration settings for Azure Active Directory Domain Services backup policy.
        - **enhanced** - (Optional) If set to `true` the enhanced backup policy will be deployed additionally to the default backup policy. This is required for the domain controller to use the enhanced backup and TrustedLaunch feature.
      - **network** - (Optional) Network configuration settings for the domain. Currently not used.
      - **network_security_rules** - (Optional) Configuration for network security rules that will be applied to all child domains in the same forest.
        - **custom_prefixes** - (Optional) An object with key value pairs that can be used as custom variables in a custom NSG template. Read the network security documentation for more details.
        - **remote_address_prefixes** - (Optional) A list of remote address prefixes in CIDR notation domain controller need to communicate with.
      - **recovery_service** - (Optional) Recovery services configuration object
        - **immutability** - (Optional) For the first initial deployment set to `disabled`. After you are happy with the deployment you should change this to `unlocked`.
    - **domains** - (Optional) A list of domain objects to configure domains and domain controller within the forest configuration.
      - **config** - (Optional) Configuration options object for domain specific settings
        - **domain_controllers** - (Optional) List of objects to configure domain controller Azure virutal machines for this domain.
          - **azure_update** - (Optional) Configuration settings for Azure Update management on the domain controller virtual machines.
            - **enabled** - (Optional) Enable Azure Update management on the domain controller.
            - **template** - (Optional) The maintenance template defined in maintenance_templates to use for patching the domain controller. Reference by name.
          - **computer_name** - (Optional) Use this setting to set a name that is set to the Windows Server. This can be helpful for the customer to align with their naming schema.
          - **dns_servers** - (Optional) A custom list of DNS Server to set on the domain controller VM network interface.
          - **enable_boot_diagnostics** - (Optional) Enable boot diagnostics for the domain controller. This is useful for troubleshooting purposes.
          - **enabled** - (Optional) Toggle to control whether resources should be deployed or not. If you set this to false after you have successfully deployed, this will destroy the resources!
          - **image_sku** - (Optional)  The image_sku that should be used for the Azure VM
          - **meta_data** - (Optional) Metadata settings for the domain controller VM.
            - **tags** - (Optional) A map of tags to assign to the domain controller VM.
          - **name** - (Optional) Custom name for the Azure VM resource. Normally this setting is not required. Only use if advised to do so.
          - **role_assignments** - (Optional) A list of role assignments to assign to the domain controller VM
            - **principal_id** - (Required) The principal ID of the user, group, or service principal to which the role assignment will be applied.
            - **role_definition_name** - (Required) The name of the role definition to assign to the principal.
          - **size** - (Optional)  The size of the Azure VM to deploy as domain controller.
          - **use_enhanced_backup** - (Optional) Enable enhanced backup for the domain controller. This requires the forest backup policy to have enhanced backup enabled as well.
          - **use_trusted_launch** - (Optional) Enable trusted launch for the domain controller VM. Trusted launch provides enhanced security features.
        - **domain_type** - (Optional) The type of the domain. Possible values are `_root` and `_child`.
        - **network** - (Optional) Network configuration settings for the domain. Currently not used.
        - **network_security_rules** - (Optional) Configuration for network security rules that will be applied only to this domain
          - **custom_prefixes** - (Optional)  An object with key value pairs that can be used as custom variables in a custom NSG template. Read the network security documentation for more details.
          - **remote_address_prefixes** - (Optional) A list of remote address prefixes in CIDR notation domain controller need to communicate with.
        - **optional_routes** - (Optional) A list of optional tags that should be added to the domain resources to route directly to the internet.
        - **parent_domain_name** - (Optional) The FQDN of the parent domain for child domains.
        - **users** - (Optional) User configuration settings for the domain.
          - **customer_role_assignment** - (Optional) List of user names defined in the core configuration to assign some general permissions to the domain controller for this domain. Only use the username not the upn.
      - **fqdn** - (Optional) The fully qualified domain name for the domain within the forest.
      - **index** - (Optional) The index of the domain within the forest. Used to create unique resource names. Start with `1` and increment for each additional domain.
      - **netbios_name** - (Optional) The NetBIOS name for the domain within the forest.
    - **enabled** - (Optional)  Toggle to control whether resources should be deployed or not. If you set this to false after you have successfully deployed, this will destroy the resources!
    - **fqdn** - (Optional) The fully qualified domain name for the domain within the forest.
    - **index** - (Optional) The index of the domain within the forest. Used to create unique resource names. Start with `1` and increment for each additional domain.
    - **netbios_name** - (Optional) The NetBIOS name for the domain within the forest.
  - **maintenance_templates** - (Optional) Use this block to define maintenance templates for in-guest patching of domain controllers. The template will not deploy anything itself, but can be referenced in the domain controller configuration to enable Azure Update management.
    - **classifications_to_include** - (Optional) A list of classifications to include in the maintenance template.
    - **duration** - (Optional) The duration of the maintenance window in HH:MM format.
    - **expiration_date_time** - (Optional) The expiration date and time of the maintenance template in ISO 8601 format.
    - **in_guest_user_patch_mode** - (Optional) Specifies the in-guest user patch mode for the maintenance template. Possible values are "User" and "System".
    - **name** - (Required) The name of the maintenance template. This name is used to reference the template in domain controller configurations.
    - **reboot** - (Optional) Specifies whether a reboot is required after applying the maintenance template.
    - **recur_every** - (Required) Specifies the recurrence interval for the maintenance template.
    - **scope** - (Optional) The scope of the maintenance template.
    - **start_date_time** - (Required) The start date and time of the maintenance template in ISO 8601 format.
    - **time_zone** - (Optional) The time zone for the maintenance template.
    - **visibility** - (Optional) The visibility setting for the maintenance template.
- **tags** - (Optional) A map of tags to assign to all resources deployed.

<!-- /MARINATED: configure\_adds\_resources -->

### Advanced Settings

The `advanced` setting allows configuration of additional settings that are not covered by the main configuration schema.

The following resources currently support advanced configuration:

- azurerm\_virtual\_network
- azurerm\_subnet
- azurerm\_resource\_group
- azurerm\_route\_table
- azurerm\_subnet
- azurerm\_key\_vault
- azurerm\_windows\_virtual\_machine
- azurerm\_network\_interface
- azurerm\_recovery\_services\_vault
- azurerm\_backup\_policy\_vm

References to `mode`can be one of the following:

- `production`
- `firetest`
- `emergency`

#### azurerm\_resource\_group

- azurerm\_resource\_group\\[`%mode%`\\]\\[`%forest%`\\].name

#### azurerm\_virtual\_network

- azurerm\_virtual\_network\\[`%mode%`\\]\\[`%forest%`\\].name
- azurerm\_virtual\_network\\[`%forest%`\\].address\_space

#### azurerm\_subnet

- azurerm\_subnet\\[`%forest%`\\]\\[`%domain%`\\].address\_prefix
- azurerm\_subnet\\[`%mode%`\\]\\[key\\].private\_endpoint\_network\_policies\_enabled
- azurerm\_subnet\\[`%mode%`\\]\\[key\\].private\_link\_service\_network\_policies\_enabled
- azurerm\_subnet\\[`%mode%`\\]\\[key\\].service\_endpoints
- azurerm\_subnet\\[`%mode%`\\]\\[key\\].service\_endpoint\_policy\_ids
- azurerm\_subnet\\[`%mode%`\\]\\[key\\].delegation
- azurerm\_subnet\\[`%mode%`\\]\\["AzureBastionSubnet"\\].private\_endpoint\_network\_policies\_enabled
- azurerm\_subnet\\[`%mode%`\\]\\["AzureBastionSubnet"\\].private\_link\_service\_network\_policies\_enabled
- azurerm\_subnet\\[`%mode%`\\]\\["AzureBastionSubnet"\\].service\_endpoints
- azurerm\_subnet\\[`%mode%`\\]\\["AzureBastionSubnet"\\].service\_endpoint\_policy\_ids
- azurerm\_subnet\\[`%mode%`\\]\\["AzureBastionSubnet"\\].delegation

#### azurerm\_key\_vault

- azurerm\_key\_vault\\["production"\\]\\[`%forest%`\\].name

#### azurerm\_windows\_virtual\_machine

- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].name
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].admin\_username
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].admin\_password
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].os\_disk.storage\_account\_type
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].source\_image\_reference.publisher
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].source\_image\_reference.offer
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].source\_image\_reference.version
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].zone
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].enable\_automatic\_updates
- azurerm\_windows\_virtual\_machine\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].provision\_vm\_agent

#### azurerm\_network\_interface

- azurerm\_network\_interface\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].ip\_configuration\\[0\\].private\_ip\_address\_allocation
- azurerm\_network\_interface\\[`%mode%`\\]\\[`%forest%`\\]\\[`%domain%`\\]\\[`%computer_name%`\\].ip\_configuration\\[0\\].ip\_address

#### azurerm\_recovery\_services\_vault

- azurerm\_recovery\_services\_vault\\[`%forest%`\\].name
- azurerm\_recovery\_services\_vault\\[`%forest%`\\].sku
- azurerm\_recovery\_services\_vault\\[`%forest%`\\].storage\_mode\_type
- azurerm\_recovery\_services\_vault\\[`%forest%`\\].identity.type

#### azurerm\_backup\_policy\_vm

- azurerm\_backup\_policy\_vm\\[`%forest%`\\].policy\_type
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].instant\_restore\_retention\_days
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].timezone
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].backup.frequency
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].backup.time
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].retention\_daily
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].retention\_weekly
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].retention\_monthly
- azurerm\_backup\_policy\_vm\\[`%forest%`\\].retention\_yearly

Type:

```hcl
object({
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
              data_disk_size : optional(number, 20)
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
```

Default: `{}`

---

Generated by MarinateMD
