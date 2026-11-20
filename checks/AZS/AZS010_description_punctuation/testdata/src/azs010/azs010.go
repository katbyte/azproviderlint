package azs010

import (
	apischema "github.com/hashicorp/go-azure-sdk/resource-manager/apimanagement/2022-08-01/schema"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// the provider reaches the plugin sdk through aliases like this one
type Schema = schema.Schema

const shared = "Shared by several properties"

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "manages a Thing", // want `description should start with a capital letter` `description should end with a full stop`

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Thing .", // want `description has a space before '\.'`
			},

			"alias": {
				Type:        schema.TypeString,
				Description: " The alias of the Thing. ", // want `description should not start with a space` `description should not end with a space`
			},

			"enabled": {
				Type:        schema.TypeBool,
				Description: "Should the Thing be enabled.", // want `description asks a question, so it should end with a question mark`
			},

			"public": {
				Type:        schema.TypeBool,
				Description: "Is the Thing public", // want `description asks a question, so it should end with a question mark`
			},

			"mode": {
				Type:        schema.TypeString,
				Description: "The mode of the Thing.  Possible values are `Fast`,`Slow` and `Off` . defaults to `Off`..", // want `description has more than one space in a row` `description has no space after ','` `description has a space before '\.'` `sentence should start with a capital letter` `description has doubled punctuation "\.\."`
			},

			"note": {
				Type:        schema.TypeString,
				Description: "A note ( free text ), eg \"hello\" , etc", // want `description has a space just inside a bracket` `description has a space just inside a bracket` `"eg" should be "e\.g\."` `description has a space before ','` `"etc" should be "etc\."`
			},

			"version": {
				Type:        schema.TypeString,
				Description: "The version, one of `v1` or `v2", // want `description has a backtick with no partner` `description should end with a full stop`
			},

			"owner": {
				Type:        schema.TypeString,
				Description: "The owner (a user or a group.", // want `description has a '\(' that is never closed`
			},

			"label": {
				Type:        schema.TypeString,
				Description: "The “label” – it’s shown in the portal…", // want `description has a curly quote '“' - use "\\""` `description has a curly quote '”' - use "\\""` `description has an en dash '–' - use "-"` `description has a curly quote '’' - use "'"` `description has a single-character ellipsis '…' - use "\.\.\."`
			},

			"separator": {
				Type:        schema.TypeString,
				Description: "The paths to skip, separated by ;", // want `description ends with a ';' on its own - if it is a value, put it in backticks` `description should end with a full stop`
			},

			"link": {
				Type:        schema.TypeString,
				Description: "See [the docs](https://example.com/things(v=1))", // want `description should end with a full stop`
			},

			"raw": {
				Type:        schema.TypeString,
				Description: `the raw one`, // want `description should start with a capital letter` `description should end with a full stop`
			},

			"shared": {
				Type:        schema.TypeString,
				Description: shared, // want `description should end with a full stop`
			},

			"joined": {
				Type:        schema.TypeString,
				Description: "The first half, " + "and the second", // want `description should end with a full stop`
			},

			"zoned": {
				Type:        schema.TypeBool,
				Description: "Should zones be used. Changing this forces a new Thing to be created.", // want `description asks a question, so it should end with a question mark`
			},

			"bullet": {
				Type:        schema.TypeString,
				Description: "- The bullet\tleft over. For example:`a`or `b `.", // want `description starts with "-" - start with its first word` `description has a tab in it - use a space` `description has no space after ':'` `description has no space after the code` `description has a space just inside a backtick`
			},

			"aliased": &Schema{
				Type:        schema.TypeString,
				Description: "the aliased one.", // want `description should start with a capital letter`
			},
		},
	}
}

func flaggedFramework() fwschema.Schema {
	return fwschema.Schema{
		MarkdownDescription: "Manages a Thing", // want `description should end with a full stop`

		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{
				Description:         "the name of the Thing.", // want `description should start with a capital letter`
				MarkdownDescription: "The name of the Thing",  // want `description should end with a full stop`
			},
		},
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "Manages a Thing" // want `description should end with a full stop`
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Thing. Changing this forces a new Thing to be created.",
			},

			"enabled": {
				Type:        schema.TypeBool,
				Description: "Should the Thing be enabled? Defaults to `true`.",
			},

			"sku_name": {
				Type:        schema.TypeString,
				Description: "`sku_name` is two parts joined by an underscore(\\_), e.g. `Developer_1`. See [the docs](https://example.com/sku.html) for more.",
			},

			"runtime": {
				Type:        schema.TypeString,
				Description: "The version of .NET to use, i.e. the one the app targets (such as 8.0, 9.0, etc.). iSCSI is not supported...",
			},

			"zone": {
				Type:        schema.TypeString,
				Description: "The zone, in the form `region:zone`, at 10:30 UTC or 1,000 ms later: a) the first, b) the second.",
			},

			"blank": {
				Type:        schema.TypeString,
				Description: "",
			},

			"built": {
				Type:        schema.TypeString,
				Description: describe("built"),
			},
		},
	}
}

func describe(name string) string { return "the " + name }

// an Azure API model's description is the user's own text
func model() apischema.SchemaContractProperties {
	return apischema.SchemaContractProperties{Description: "whatever the user wrote"}
}

type notAResource struct{}

func (notAResource) Description() string { return "describes something else" }
