package azs020

import (
	"github.com/hashicorp/go-azure-sdk/backups"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var _ = backups.ActionStopProtection

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing. Use `example_thingy` for the older kind.", // want `description names the resource "example_thingy" - the provider has no resource called that`

		Schema: map[string]*schema.Schema{
			"client_certificate_enabled": {
				Type:        schema.TypeBool,
				Description: "Whether client certificates are enabled.",
			},

			// the property was renamed and the description was not
			"client_certificate_mode": {
				Type:        schema.TypeString,
				Description: "The client certificate mode. This has no effect when `client_cert_enabled` is `false`.", // want `description names "client_cert_enabled" - nothing in the provider is called that`
			},

			"site_config": {
				Type: schema.TypeList,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{"linux_fx_version": {Type: schema.TypeString}}},
			},

			"runtime": {
				Type:        schema.TypeString,
				Description: "The runtime, as set in `site_config.0.linux_fx_release` or `site_config[0].linux_fx`.", // want `description names "linux_fx_release" - nothing in the provider is called that` `description names "linux_fx" - nothing in the provider is called that`
			},

			"source_id": {
				Type:        schema.TypeString,
				Description: "The ID of the source, such as `example_gadgets.main.id` or `example_widget.main.no_such_thing`.", // want `description names the resource "example_gadgets" - the provider has no resource called that`
			},

			"size": {
				Type:        schema.TypeString,
				Description: "The size. Set it in the `skew` block, not the `site_config` block.", // want `description names "skew" - nothing in the provider is called that`
			},

			"label": helperSchema("The label. Conflicts with `label_suffix`."), // want `description names "label_suffix" - nothing in the provider is called that`
		},
	}
}

func helperSchema(description string) *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Description: description}
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing. Use `example_gadget` or `example_widget` with it, or any of `example_thing_*`.",

		Schema: map[string]*schema.Schema{
			"client_certificate_mode": {
				Type:        schema.TypeString,
				Description: "The client certificate mode. This has no effect when `client_certificate_enabled` is `false`.",
			},

			"runtime": {
				Type:        schema.TypeString,
				Description: "The runtime, as set in `site_config.0.linux_fx_version` or `site_config[0].linux_fx_version`.",
			},

			// a property of another service's resource, which this package does not import
			"scopes": {
				Type:        schema.TypeString,
				Description: "The scopes, such as a Widget's `dapr.app_id` or a Gadget's `versionless_id`.",
			},

			// values: one the SDK defines, one word, and ones that are not names at all
			"action": {
				Type:        schema.TypeString,
				Description: "The action to take, such as `suspend_backups`, `enabled`, `Standard_LRS`, `application/json` or `a-b`.",
			},

			"tags": {
				Type:        schema.TypeString,
				Description: "The tags. Use `ignore_changes` to leave them alone, or `count.index` to number them.",
			},

			"mode": {
				Type:        schema.TypeString,
				Description: "The mode, `read_only` or `read_write`.",
			},
		},
	}
}

var modes = []string{"read_only", "read_write"}
