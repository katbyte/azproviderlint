# AZS001 - typed SDK model numeric fields must be 64-bit (int64/float64)

AZS001 reports number fields in typed SDK models (struct fields tagged `tfschema`) that are not 64-bit: `int`, `int16`, or `int32` instead of `int64`, and `float32` instead of `float64`.

The typed SDK's `Encode` and `Decode` only handle `int64` and `float64`. Any other width compiles fine and then fails at runtime.

Slices, maps, and pointers of these types are checked too. Named types are resolved to what they really are, so `type Capacity int` is also reported.

## Flagged Code

```go
type ServerModel struct {
	Capacity     int       `tfschema:"capacity"`
	Priority     int32     `tfschema:"priority"`
	CpuThreshold float32   `tfschema:"cpu_threshold"`
	AllowedPorts []int     `tfschema:"allowed_ports"`
}
```

## Passing Code

```go
type ServerModel struct {
	Capacity     int64     `tfschema:"capacity"`
	Priority     int64     `tfschema:"priority"`
	CpuThreshold float64   `tfschema:"cpu_threshold"`
	AllowedPorts []int64   `tfschema:"allowed_ports"`
}
```

## Ignoring Reports

Put `//azignore:AZS001 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
Capacity int `tfschema:"capacity"` //azignore:AZS001 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
