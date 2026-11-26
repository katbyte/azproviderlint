# AZS021 - a MaxItems: 1 block with one property must be flattened or say why it is kept

AZS021 reports a block that can be given once and holds a single property.

The block only adds a level of nesting: the property can sit where the block does. A block kept because more properties are on the way says so in a comment.

## Flagged Code

```go
"certificate": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"thumbprint": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
		},
	},
},
```

```text
block "certificate" can be given once and holds only "thumbprint" - move it out of the block, or say why the block is kept in a comment above it
```

## Passing Code

```go
"certificate_thumbprint": {
	Type:     pluginsdk.TypeString,
	Optional: true,
},
```

```go
// the service team has said more settings are coming to this block
"certificate": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"thumbprint": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
		},
	},
},
```

## What is checked

A plugin SDK list or set with `MaxItems: 1`, whose `Elem` is a `pluginsdk.Resource` with its properties written out.

The comment is on a line of its own, right above the block. A comment inside the block is about the field under it and does not count, nor does one at the end of a line, nor a note for a linter such as `// lintignore:XS003`.

These are not checked:

- a block whose properties come from a function
- a framework attribute or block
- a schema in a test file, which is never shipped
- anything in a package named `migration` or `migrations`, which is what old state looked like

Flattening a block that has shipped is a breaking change. Until the next major version, give it the comment.

## Fix

None is offered. Move the property out of the block, or write the comment.

## Ignoring Reports

Put `//azignore:AZS021 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"certificate": { //azignore:AZS021 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
