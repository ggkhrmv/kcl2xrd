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
| `[T]` | `array` | `tags: [str]` |
| `{K:V}` | `object` | `labels: {str:str}` |
| `{any:any}` | `object` + `x-kubernetes-preserve-unknown-fields` | `config: {any:any}` |

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

- `__xrd_kind` - Schema to convert
- `__xrd_group` - API group
- `__xrd_version` - API version (default: v1alpha1)
- `__xrd_categories` - Categories list
- `__xrd_served` - Served flag (True/False)
- `__xrd_referenceable` - Referenceable flag (True/False)
- `__xrd_printer_columns` - Printer columns list

## CLI Options

- `-i, --input`: Input KCL file (required)
- `-g, --group`: API group (optional if `__xrd_group` in file)
- `-o, --output`: Output file (stdout if not specified)
- `--with-claims`: Generate claimable XRD with automatic X-prefix handling
- `--schema`: Select specific schema
- `--version`: API version (default: v1alpha1)
- `--categories`: Override categories
- `--printer-columns`: Override printer columns

## Examples

See [`examples/`](examples/) directory:

1. **postgresql.k** - Basic schema with optional fields
2. **validated.k** - Validation annotations
3. **nested-schema.k** - Nested schema references
4. **dynatrace-with-metadata.k** - Full in-file metadata
5. **preserve-unknown-fields.k** - Arbitrary properties with `{any:any}`

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
