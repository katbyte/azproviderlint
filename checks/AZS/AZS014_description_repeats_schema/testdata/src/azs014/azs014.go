package azs014

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const shared = "The shared one. Defaults to `a`."

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "(Resource) Manages a Thing, (Optional) and more.", // want `description has the tag "\(Optional\)" - the schema already says so`

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "(Required) The name of the Thing. Changing this forces a new Thing to be created.", // want `description has the tag "\(Required\)" - the schema already says so` `description says a change forces a new resource - the schema's ForceNew already does`
			},

			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Fast",
				ValidateFunc: validation.StringInSlice([]string{"Fast", "Slow"}, false),
				Description:  "The mode of the Thing. Possible values are `Fast` and `Slow`. Defaults to `Fast`.", // want `description lists the values it can take - the schema's validation already does` `description gives the default - the schema's Default already does`
			},

			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "(Boolean) Should the Thing be enabled? Possible values are `true` and `false`. Defaults to `true`.", // want `description has the tag "\(Boolean\)" - the schema already says so` `description lists the values it can take - a bool is always true or false` `description gives the default - the schema's Default already does`
			},

			"size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
				Description:  "The size of the Thing. Must be between `1` and `10`.", // want `description lists the values it can take - the schema's validation already does`
			},

			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The secret of the Thing (Optional, Sensitive).", // want `description has the tag "\(Optional, Sensitive\)" - the schema already says so`
			},

			"zones": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"1", "2", "3"}, false),
				},
				Description: "The zones to use. Valid values include `1`, `2` and `3`.", // want `description lists the values it can take - the schema's validation already does`
			},

			"sku": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"tier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tier": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "Basic",
							Description: "Defaults to `Basic`. The tier of the SKU.", // want `description gives the default - the schema's Default already does`
						},
					},
				},
				Description: "A `sku` block as defined below. Conflicts with `tier`.", // want `description points at the docs page with "as defined below" - a description has no page around it` `description says what it conflicts with - the schema's ConflictsWith or ExactlyOneOf already does`
			},

			"only": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"a", "b"}, false),
				Description:  "Possible values are `a` and `b`.", // want `description lists the values it can take - the schema's validation already does`
			},

			"raw": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "a",
				Description: `The raw one. The default is "a".`, // want `description gives the default - the schema's Default already does`
			},

			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The label of the Thing. Changing this forces a new resource.", // want `description says a change forces a new resource - the schema's ForceNew already does`
			},

			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Http", "Https"}, false),
				Description:  "The protocol to use. Can be `Http` or `Https`.", // want `description lists the values it can take - the schema's validation already does`
			},

			"verb": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"GET", "POST", "PUT"}, false),
				Description:  "Which HTTP verb to use. Options are 'GET', 'POST', and 'PUT'.", // want `description lists the values it can take - the schema's validation already does`
			},

			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "none",
				Description: "The owner, a user or a group, which defaults to `none`.", // want `description gives the default - the schema's Default already does`
			},

			"site": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Default-First-Site-Name",
				Description: "The site to use (defaults to 'Default-First-Site-Name') for discovery. If blank, defaults to 'Default-First-Site-Name'.", // want `description gives the default - the schema's Default already does` `description gives the default - the schema's Default already does`
			},

			"key": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"key", "secret"},
				Description:  "The key to sign with. Exactly one of `key` or `secret` must be specified.", // want `description says one of several must be set - the schema's ExactlyOneOf or AtLeastOneOf already does`
			},

			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
				Description: "The rules to apply. Each `rule` block supports fields defined below. A maximum of `100` `rule` blocks may be defined.", // want `description points at the docs page - a description has no page around it` `description says how many there can be - the schema's MaxItems or MinItems already does`
			},

			"title": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 80),
				Description:  "The title of the Thing. Must be between 1 and 80 characters.", // want `description says how long it can be - the schema's validation already does`
			},

			"alias": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 80),
				Description:  "The alias of the Thing, letters and dashes up to 80 characters in length.", // want `description says how long it can be - the schema's validation already does`
			},

			"repository_url": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"repository_branch", "repository_token"},
				Description:  "The repository to deploy from. `repository_branch` and `repository_token` must also be set.", // want `description says what else must be set - the schema's RequiredWith already does`
			},

			"client_secret": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"client_secret", "client_secret_setting_name"},
				Description:  "The client secret. Cannot be used with `client_secret_setting_name`.", // want `description says what it conflicts with - the schema's ConflictsWith or ExactlyOneOf already does`
			},

			"tls": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1.2", "1.3"}, false),
				Description:  "The minimum version of TLS for the site Possible values include: `1.2` and `1.3`.", // want `description lists the values it can take - the schema's validation already does`
			},

			"runbook_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Graph", "Script"}, false),
				Description:  "The type of the runbook - can be either `Graph` or `Script`.", // want `description lists the values it can take - the schema's validation already does`
			},

			"shared": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "a",
				Description: shared, // want `description gives the default - the schema's Default already does`
			},
		},
	}
}

func flaggedFramework() fwschema.Schema {
	return fwschema.Schema{
		MarkdownDescription: "Manages a Thing.",

		Attributes: map[string]fwschema.Attribute{
			"kind": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The kind of Thing. Possible values are `a` and `b`. Changing this forces a new resource to be created.", // want `description lists the values it can take - the schema's validation already does` `description says a change forces a new resource - the schema's ForceNew already does`
				Validators:          []any{stringvalidator.OneOf("a", "b")},
				PlanModifiers:       []any{stringplanmodifier.RequiresReplace()},
			},

			"on": fwschema.BoolAttribute{
				Optional:    true,
				Description: "Is the Thing on? Possible values are `true` or `false`.", // want `description lists the values it can take - a bool is always true or false`
			},

			"title": fwschema.StringAttribute{
				Optional:    true,
				Description: "The title of the Thing. Cannot exceed 80 characters in length.", // want `description says how long it can be - the schema's validation already does`
				Validators:  []any{stringvalidator.LengthBetween(1, 80)},
			},
		},

		Blocks: map[string]fwschema.Block{
			"filter": fwschema.ListNestedBlock{
				Description: "The filters to apply. Can support up to `5` `filter` blocks.", // want `description says how many there can be - the schema's MaxItems or MinItems already does`
				Validators:  []any{listvalidator.SizeAtMost(5)},
			},

			"rule": fwschema.ListNestedBlock{
				Description: "One or more `rule` blocks as defined below.", // want `description points at the docs page with "as defined below" - a description has no page around it`
			},
		},
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "Manages a Thing as documented below." // want `description points at the docs page with "as documented below" - a description has no page around it`
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing.",

		Schema: map[string]*schema.Schema{
			// nothing in the schema gives the default, the values or the rule, so the description is the only place they are written down
			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The mode of the Thing. Possible values are `Fast` and `Slow`. Defaults to `Fast`. Changing this forces a new Thing to be created. Conflicts with `speed`.",
			},

			"plain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The plain one, which defaults to `a`. Can be `a` or `b`. Changing this forces a new resource. Must be between 1 and 80 characters. Exactly one of `plain` or `mode` must be specified.",
			},

			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
				Description: "The rules to apply. A maximum of `100` `rule` blocks may be defined.",
			},

			// said in a way that tells the reader more than the schema does
			"owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				MaxItems:     1,
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The owner, a user (optional) or a group, who by default is nobody. Limit it to 64 characters to avoid collisions. Only one owner can be set unless `mode` is `Fast`.",
			},

			"sku": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tier": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"a", "b"}, false),
							Description:  "The tier of the SKU.",
						},
					},
				},
				Description: "The SKU of the Thing. Possible values are whatever the region offers.",
			},

			"built": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "a",
				Description: describe("built"),
			},
		},
	}
}

func passingFramework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"kind": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The kind of Thing. Possible values are `a` and `b`. Changing this forces a new resource to be created.",
				Validators:          []any{stringvalidator.LengthAtLeast(1)},
				PlanModifiers:       []any{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func describe(name string) string { return "the " + name + ". Defaults to `a`." }
