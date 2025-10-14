# kcl2xrd

Convert KCL (KCL Configuration Language) schemas to Crossplane Composite Resource Definitions (XRDs).

Inspired by [crossplane-tools](https://github.com/crossplane/crossplane-tools) and [kcl-openapi](https://github.com/kcl-lang/kcl-openapi).

## Quick Start

```bash
# Build from source
git clone https://github.com/ggkhrmv/kcl2xrd.git
cd kcl2xrd && make build

# Simple conversion with in-file metadata
./bin/kcl2xrd -i examples/kcl/dynatrace-with-metadata.k -o output.yaml

# Override with CLI flags
./bin/kcl2xrd -i examples/kcl/postgresql.k -g database.example.org -o output.yaml

# Generate with claims support
./bin/kcl2xrd -i examples/kcl/postgresql.k -g database.example.org --with-claims -o output.yaml
```

## Key Features

- **In-file XRD metadata** with `__xrd_` prefix variables - define everything in your KCL files
- **`@xrd` annotation** - mark parent schema, ignore unrelated code
- **Validation annotations** - patterns, enums, ranges, string/numeric constraints, CEL expressions
- **Kubernetes-specific annotations** - immutability, preserveUnknownFields, mapType, listType, listMapKeys
- **Nested schema expansion** - automatic reference resolution
- **`{any:any}` syntax** - arbitrary property objects with `@preserveUnknownFields`
- **Claims support** - automatic X-prefix handling for composite resources

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/ggkhrmv/kcl2xrd/releases):

```bash
# Linux (AMD64)
curl -LO https://github.com/ggkhrmv/kcl2xrd/releases/latest/download/kcl2xrd-linux-amd64
chmod +x kcl2xrd-linux-amd64 && sudo mv kcl2xrd-linux-amd64 /usr/local/bin/kcl2xrd

# macOS (Intel)
curl -LO https://github.com/ggkhrmv/kcl2xrd/releases/latest/download/kcl2xrd-darwin-amd64
chmod +x kcl2xrd-darwin-amd64 && sudo mv kcl2xrd-darwin-amd64 /usr/local/bin/kcl2xrd

# macOS (Apple Silicon)
curl -LO https://github.com/ggkhrmv/kcl2xrd/releases/latest/download/kcl2xrd-darwin-arm64
chmod +x kcl2xrd-darwin-arm64 && sudo mv kcl2xrd-darwin-arm64 /usr/local/bin/kcl2xrd
```

### From Source

```bash
go install github.com/ggkhrmv/kcl2xrd/cmd/kcl2xrd@latest
# or
git clone https://github.com/ggkhrmv/kcl2xrd.git && cd kcl2xrd && make build
```

## Usage

### Basic Conversion

```kcl
# postgresql.k
schema PostgreSQLInstance:
    # Storage in GB (required)
    storageGB: int
    
    # Instance size  
    instanceSize?: str = "small"
```

```bash
kcl2xrd -i postgresql.k -g database.example.org -o postgresql.yaml
```

### With In-File Metadata

```kcl
# Full automation - no CLI flags needed
__xrd_kind = "DynatraceAlerting"
__xrd_group = "monitoring.crossplane.io"
__xrd_categories = ["monitoring", "alerting"]

# @xrd
schema DynatraceAlerting:
    name: str
    config: {str:str}
```

```bash
kcl2xrd -i file.k -o output.yaml
```

### Claims Support

When using `--with-claims`, the tool automatically handles X-prefix naming:

```bash
# Schema: PostgreSQLInstance (no X-prefix)
kcl2xrd -i postgresql.k -g db.example.org --with-claims -o output.yaml

# Generates:
# - XRD Kind: XPostgreSQLInstance (X-prefix added)
# - Claim Kind: PostgreSQLInstance (original name)
```

If schema already has X-prefix:

```bash
# Schema: XDatabase (has X-prefix)
kcl2xrd -i xdatabase.k -g db.example.org --with-claims -o output.yaml

# Generates:
# - XRD Kind: XDatabase (keeps X-prefix)
# - Claim Kind: Database (X-prefix removed)
```

## Type Mappings

| KCL Type | OpenAPI Type | Example |
|----------|---------------|---------|
| `str` | `string` | `name: str` |
| `int` | `integer` | `count: int` |
| `float` | `number` | `price: float` |
| `bool` | `boolean` | `enabled: bool` |
| `any` | (no type) + `x-kubernetes-preserve-unknown-fields` | `principal?: any` |
| `[T]` | `array` | `tags: [str]` |
| `{K:V}` | `object` | `labels: {str:str}` |
| `{any:any}` | `object` + `x-kubernetes-preserve-unknown-fields` | `config: {any:any}` |

**Note:** The `any` type is particularly useful for fields that can accept arbitrary JSON/YAML data (like AWS IAM policy principals, actions, etc.). When using `any` type with `@preserveUnknownFields` annotation, the field will not have a type constraint, allowing maximum flexibility.

## Annotations Reference

### Schema-Level Annotations

#### `@xrd`
Marks the schema to be converted to XRD. Only one schema in a file should be marked with `@xrd`.

```kcl
# @xrd
schema MyResource:
    name: str
```

### String Validation Annotations

#### `@pattern(regex)`
Applies a regex pattern validation to string fields.

```kcl
# @pattern("^[a-z0-9-]+$")
name: str
```

#### `@minLength(n)`
Sets minimum length for string fields.

```kcl
# @minLength(3)
name: str
```

#### `@maxLength(n)`
Sets maximum length for string fields.

```kcl
# @maxLength(63)
name: str
```

### Numeric Validation Annotations

#### `@minimum(n)`
Sets minimum value for integer fields.

```kcl
# @minimum(0)
replicas: int
```

#### `@maximum(n)`
Sets maximum value for integer fields.

```kcl
# @maximum(100)
replicas: int
```

### Enum Validation

#### `@enum([values])`
Restricts field to specific allowed values.

```kcl
# @enum(["active", "inactive", "pending"])
status: str
```

### Kubernetes-Specific Annotations

#### `@immutable`
Marks a field as immutable (sets `x-kubernetes-immutable: true`).

```kcl
# @immutable
resourceId: str
```

#### `@preserveUnknownFields`
Allows arbitrary properties (sets `x-kubernetes-preserve-unknown-fields: true`). Typically used with `{any:any}` type.

```kcl
# @preserveUnknownFields
config: {any:any}
```

#### `@mapType(type)`
Sets `x-kubernetes-map-type`. Valid values: `"atomic"`, `"granular"`.

```kcl
# @mapType("atomic")
settings: {str:str}
```

#### `@listType(type)`
Sets `x-kubernetes-list-type`. Valid values: `"atomic"`, `"set"`, `"map"`.

```kcl
# @listType("set")
tags: [str]
```

#### `@listMapKeys([keys])`
Sets `x-kubernetes-list-map-keys` for list-map type lists.

```kcl
# @listType("map")
# @listMapKeys(["name"])
items: [Item]
```

### CEL Validation

#### `@validate(rule, message?)`
Adds CEL (Common Expression Language) validation rules with optional error message.

```kcl
# @validate("self > 0", "Must be positive")
value: int

# @validate("self.startsWith('prefix-')")
identifier: str
```

### Complete Example

```kcl
schema ValidatedResource:
    # String with pattern and length constraints
    # @pattern("^[a-z0-9-]+$")
    # @minLength(3)
    # @maxLength(63)
    name: str
    
    # Enum validation
    # @enum(["active", "inactive", "pending"])
    status?: str = "active"
    
    # Immutable field
    # @immutable
    resourceId: str
    
    # Numeric constraints
    # @minimum(0)
    # @maximum(100)
    replicas?: int = 1
    
    # Arbitrary properties with atomic map type
    # @preserveUnknownFields
    # @mapType("atomic")
    settings?: {any:any}
    
    # List with set semantics
    # @listType("set")
    tags?: [str]
    
    # CEL validation
    # @validate("self > 0", "Must be positive")
    value: int
```

## Metadata Variables

Define in your KCL file with `__xrd_` prefix:

- `__xrd_kind` - Schema to convert (string literal)
- `__xrd_group` - API group (string literal or format expression)
- `__xrd_version` - API version (default: v1alpha1)
- `__xrd_categories` - Categories list
- `__xrd_served` - Served flag (True/False)
- `__xrd_referenceable` - Referenceable flag (True/False)
- `__xrd_printer_columns` - Printer columns list

**Note on `__xrd_group`:** 
- String literals (e.g., `__xrd_group = "example.org"`) are extracted automatically
- Format expressions and any KCL expressions (e.g., `__xrd_group = "{}.{}".format(var1, var2)` or `__xrd_group = "{}.{}".format(_xrSubgroup, settings.PLATFORM_API_GROUP)`) are automatically evaluated using the KCL runtime
- This provides maximum flexibility - you can use any valid KCL expression to compute the group

**Note on `__xrd_kind`:**
- Specifies the `spec.names.kind` for the XRD (e.g., `"XBucket"`, `"Database"`)
- The XRD `metadata.name` will use the plural of this kind (e.g., `"xbuckets.group"`, `"databases.group"`)
- If not specified, defaults to the schema name
- Useful when you want the XRD kind to differ from the schema name

### Example with Format Expressions

```kcl
# Define variables
_xrSubgroup = "aws"
_platformGroup = "mycorp.io"

# Use format expression - automatically evaluated by KCL
__xrd_kind = "XBucket"
__xrd_group = "{}.{}".format(_xrSubgroup, _platformGroup)

# @xrd
schema Bucket:
    name: str
```

```bash
# No --group flag needed - expression is automatically evaluated
kcl2xrd -i bucket.k -o bucket.yaml
# Generates:
# - metadata.name: xbuckets.aws.mycorp.io (plural of XBucket)
# - spec.names.kind: XBucket (from __xrd_kind)
# - spec.group: aws.mycorp.io (from evaluated __xrd_group)
```

### Example with Complex Expressions

```kcl
# Using property access and nested structures
settings = {
    PLATFORM_API_GROUP: "platform.example.com"
}

_xrSubgroup = "storage"

__xrd_kind = "XDatabase"
__xrd_group = "{}.{}".format(_xrSubgroup, settings.PLATFORM_API_GROUP)

# @xrd
schema Database:
    name: str
```

```bash
# All expressions are evaluated using KCL runtime
kcl2xrd -i database.k -o database.yaml
# Automatically resolves to: storage.platform.example.com
```

## CLI Options

- `-i, --input`: Input KCL file (required)
- `-g, --group`: API group (optional if `__xrd_group` in file)
- `-o, --output`: Output file (stdout if not specified)
- `--with-claims`: Generate claimable XRD with automatic X-prefix handling
- `--schema`: Select specific schema
- `--version`: API version (default: v1alpha1)
- `--categories`: Override categories
- `--printer-columns`: Override printer columns

## Best Practices

### Working with Complex KCL Files

When your KCL files contain both schema definitions and other code (like composition templates, module-level variables, etc.), follow these guidelines:

1. **Use the `@xrd` annotation** to explicitly mark which schema should be converted:

```kcl
# Other code and imports
import base.schemas as schemas

# @xrd  <-- Mark the schema for conversion
schema MyResource:
    name: str
    config?: any

# Other module-level variables (these won't be parsed as schema fields)
_composition: schemas.Composition{
    xrKind: "MyResource"
}
```

2. **Module-level variables after schemas are ignored**: The parser stops collecting schema fields when it encounters a non-indented line (module-level code). This ensures composition templates and other variables don't pollute your XRD schema.

3. **Use `any` type for flexible fields**: For fields that need to accept arbitrary JSON/YAML structures (like IAM policies, custom configurations), use the `any` type with `@preserveUnknownFields`:

```kcl
schema PolicyStatement:
    # @preserveUnknownFields
    # Can be a string, array, or object
    principal?: any
    
    # @preserveUnknownFields
    # Can be a string, array, or object
    action?: any
```

This generates fields without a `type` constraint, only with `x-kubernetes-preserve-unknown-fields: true`, allowing maximum flexibility.

## Examples

See [`examples/`](examples/) directory:

1. **postgresql.k** - Basic schema with optional fields
2. **validated.k** - Validation annotations
3. **nested-schema.k** - Nested schema references
4. **dynatrace-with-metadata.k** - Full in-file metadata
5. **preserve-unknown-fields.k** - Arbitrary properties with `{any:any}`
6. **s3-bucket-with-policy.k** - Complex example with `any` type fields and IAM policies

## Development

```bash
# Build
make build

# Run tests
make test

# Generate examples
make examples

# Create release (requires tag)
git tag v1.0.0 && git push origin v1.0.0
```

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.
