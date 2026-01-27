# This example file is from the Terraform Azure Verified Module for Storage Accounts
# There are some complex variables in this module

variable "custom_domain" {
  type = object({
    name          = string
    use_subdomain = optional(bool)
  })
  default     = null
  description = <<-EOT
 - `name` - (Required) The Custom Domain Name to use for the Storage Account, which will be validated by Azure.
 - `use_subdomain` - (Optional) Should the Custom Domain Name be validated by using indirect CNAME validation?
EOT
}

variable "local_user" {
  type = map(object({
    home_directory       = optional(string)
    name                 = string
    ssh_key_enabled      = optional(bool)
    ssh_password_enabled = optional(bool)
    permission_scope = optional(list(object({
      resource_name = string
      service       = string
      permissions = object({
        create = optional(bool)
        delete = optional(bool)
        list   = optional(bool)
        read   = optional(bool)
        write  = optional(bool)
      })
    })))
    ssh_authorized_key = optional(list(object({
      description = optional(string)
      key         = string
    })))
    timeouts = optional(object({
      create = optional(string)
      delete = optional(string)
      read   = optional(string)
      update = optional(string)
    }))
  }))
  default     = {}
  description = <<-EOT
 - `home_directory` - (Optional) The home directory of the Storage Account Local User.
 - `name` - (Required) The name which should be used for this Storage Account Local User. Changing this forces a new Storage Account Local User to be created.
 - `ssh_key_enabled` - (Optional) Specifies whether SSH Key Authentication is enabled. Defaults to `false`.
 - `ssh_password_enabled` - (Optional) Specifies whether SSH Password Authentication is enabled. Defaults to `false`.

 ---
 `permission_scope` block supports the following:
 - `resource_name` - (Required) The container name (when `service` is set to `blob`) or the file share name (when `service` is set to `file`), used by the Storage Account Local User.
 - `service` - (Required) The storage service used by this Storage Account Local User. Possible values are `blob` and `file`.

 ---
 `permissions` block supports the following:
 - `create` - (Optional) Specifies if the Local User has the create permission for this scope. Defaults to `false`.
 - `delete` - (Optional) Specifies if the Local User has the delete permission for this scope. Defaults to `false`.
 - `list` - (Optional) Specifies if the Local User has the list permission for this scope. Defaults to `false`.
 - `read` - (Optional) Specifies if the Local User has the read permission for this scope. Defaults to `false`.
 - `write` - (Optional) Specifies if the Local User has the write permission for this scope. Defaults to `false`.

 ---
 `ssh_authorized_key` block supports the following:
 - `description` - (Optional) The description of this SSH authorized key.
 - `key` - (Required) The public key value of this SSH authorized key.

 ---
 `timeouts` block supports the following:
 - `create` - (Defaults to 30 minutes) Used when creating the Storage Account Local User.
 - `delete` - (Defaults to 30 minutes) Used when deleting the Storage Account Local User.
 - `read` - (Defaults to 5 minutes) Used when retrieving the Storage Account Local User.
 - `update` - (Defaults to 30 minutes) Used when updating the Storage Account Local User.
EOT
  nullable    = false
}

variable "local_user_enabled" {
  type        = bool
  default     = false
  description = "(Optional) Should Storage Account Local Users be enabled? Defaults to `false`."
}

variable "network_rules" {
  type = object({
    bypass                     = optional(set(string), ["AzureServices"])
    default_action             = optional(string, "Deny")
    ip_rules                   = optional(set(string), [])
    virtual_network_subnet_ids = optional(set(string), [])
    private_link_access = optional(list(object({
      endpoint_resource_id = string
      endpoint_tenant_id   = optional(string)
    })))
    timeouts = optional(object({
      create = optional(string)
      delete = optional(string)
      read   = optional(string)
      update = optional(string)
    }))
  })
  default     = {}
  description = <<-EOT
 > Note the default value for this variable will block all public access to the storage account. If you want to disable all network rules, set this value to `null`.

 - `bypass` - (Optional) Specifies whether traffic is bypassed for Logging/Metrics/AzureServices. Valid options are any combination of `Logging`, `Metrics`, `AzureServices`, or `None`.
 - `default_action` - (Required) Specifies the default action of allow or deny when no other rules match. Valid options are `Deny` or `Allow`.
 - `ip_rules` - (Optional) List of public IP or IP ranges in CIDR Format. Only IPv4 addresses are allowed. Private IP address ranges (as defined in [RFC 1918](https://tools.ietf.org/html/rfc1918#section-3)) are not allowed.
 - `storage_account_id` - (Required) Specifies the ID of the storage account. Changing this forces a new resource to be created.
 - `virtual_network_subnet_ids` - (Optional) A list of virtual network subnet ids to secure the storage account.

 ---
 `private_link_access` block supports the following:
 - `endpoint_resource_id` - (Required) The resource id of the resource access rule to be granted access.
 - `endpoint_tenant_id` - (Optional) The tenant id of the resource of the resource access rule to be granted access. Defaults to the current tenant id.

 ---
 `timeouts` block supports the following:
 - `create` - (Defaults to 60 minutes) Used when creating the  Network Rules for this Storage Account.
 - `delete` - (Defaults to 60 minutes) Used when deleting the Network Rules for this Storage Account.
 - `read` - (Defaults to 5 minutes) Used when retrieving the Network Rules for this Storage Account.
 - `update` - (Defaults to 60 minutes) Used when updating the Network Rules for this Storage Account.
EOT
}

variable "sftp_enabled" {
  type        = bool
  default     = false
  description = "(Optional) Boolean, enable SFTP for the storage account.  Defaults to `false`."
}

variable "shared_access_key_enabled" {
  type        = bool
  default     = false
  description = "(Optional) Indicates whether the storage account permits requests to be authorized with the account access key via Shared Key. If false, then all requests, including shared access signatures, must be authorized with Azure Active Directory (Azure AD). The default value is `false`."
}

variable "static_website" {
  type = map(object({
    error_404_document = optional(string)
    index_document     = optional(string)
  }))
  default     = null
  description = <<-EOT
 - `error_404_document` - (Optional) The absolute path to a custom webpage that should be used when a request is made which does not correspond to an existing file.
 - `index_document` - (Optional) The webpage that Azure Storage serves for requests to the root of a website or any subfolder. For example, index.html. The value is case-sensitive.
EOT
}

variable "app_config" {
  type = object({
    database = optional(object({
      host     = string
      port     = optional(number, 5432)
      ssl_mode = optional(string, "require")
    }))
    cache = optional(object({
      redis_url = string
      ttl       = optional(number, 3600)
    }))
  })
  description = <<-EOT
<!-- MARINATED: app_config -->

- `cache` - (Optional) # TODO: Add description for cache
  - `redis_url` - (Required) # TODO: Add description for redis_url
  - `ttl` - (Optional) # TODO: Add description for ttl
- `database` - (Optional) # TODO: Add description for database
  - `host` - (Required) # TODO: Add description for host
  - `port` - (Optional) # TODO: Add description for port
  - `ssl_mode` - (Optional) # TODO: Add description for ssl_mode


<!-- /MARINATED: app_config -->
  EOT
}
