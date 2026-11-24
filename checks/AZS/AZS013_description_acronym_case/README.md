# AZS013 - resource and property descriptions must write acronyms in their usual case

AZS013 reports an acronym in a resource or property description that is not written the way it is everywhere else: `id` or `Id` for `ID`, `Api` for `API`, `url` for `URL`.

## Flagged Code

```go
"subnet_id": {
	Type:        pluginsdk.TypeString,
	Description: "The id of the subnet the Api Management Service's vms use, as a url.",
},
```

```text
"id" should be "ID"
"Api" should be "API"
"vms" should be "VMs"
"url" should be "URL"
```

## Passing Code

```go
"subnet_id": {
	Type:        pluginsdk.TypeString,
	Description: "The ID of the subnet the API Management Service's VMs use, as a URL.",
},
```

## What counts

These acronyms are checked, in any other case than the one shown:

`ID` `IP` `URL` `URI` `API` `DNS` `HTTP` `HTTPS` `JSON` `XML` `YAML` `SQL` `SSL` `TLS` `SSH` `TCP` `UDP` `VM` `CPU` `GPU` `SKU` `FQDN` `CIDR` `GUID` `UUID` `UTC` `HTML` `FTP` `FTPS` `SFTP` `SMTP` `NFS` `SMB` `VPN` `IOPS` `TTL` `ACL` `HSM` `MSI` `RBAC` `SAML` `JWT` `SDK` `CLI` `NSG` `CDN` `CORS` `CSV` `RSA` `AES` `LDAP` `LDAPS` `DHCP` `BGP` `ICMP` `SSD` `HDD` `VHD` `MFA` `SSO` `NAT` `NIC` `SAS` `LUN` `WAF` `KQL` `PEM` `PFX` `PKI` `CNAME` `AKS` `PHP` `TDE` `UPN` `SPN` `ASCII` `IPv4` `IPv6` `OAuth` `vCPU` `NVMe` `DDoS` `OData` `gRPC` `VNet`

None of them is also an everyday word. A plural counts too: `ids` should be `IDs`.

Only prose is checked. These are left alone, since there the case is part of a name or a value:

- code in backticks: `` `azurerm_subnet.example.id` ``
- web addresses, and the address a link points to
- anything inside double quotes: `"az vm list --query id"`
- a word with a dot, an underscore or an equals sign in it: `example.id`, `id_token`
- a lowercase part of a hyphenated word: `ssh-rsa`, `api-version`
- a longer word: `clientId`, `idle`

A value that happens to be an acronym belongs in backticks: `` Defaults to `http`. ``

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

## Fix

The fix rewrites the acronym in its usual case.

## Options

| Option | Default | Effect |
|---|---|---|
| `acronym` | | more acronyms to check, each written the way it should be |
| `ignore` | | acronyms from the list above to leave alone |

Both take several values: a list in the golangci settings, or the flag repeated on the command line.

```yaml
          AZS013:
            acronym: [IoT, AAD]
            ignore: [VNet]
```

```bash
azproviderlint -AZS013 -AZS013.acronym=IoT -AZS013.acronym=AAD -AZS013.ignore=VNet ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS013 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "The id of the subnet.", //azignore:AZS013 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
