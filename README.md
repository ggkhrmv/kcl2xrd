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
- **KCL runtime evaluation** - automatically evaluates metadata variables including format expressions, property access, and variable references
- **Import support** - reuse central configuration files across multiple XRDs with KCL imports
- **`@xrd` annotation** - mark parent schema, ignore unrelated code
- **Validation annotations** - patterns, enums, ranges, string/numeric constraints, CEL expressions, oneOf/anyOf schema composition
- **Kubernetes-specific annotations** - immutability, preserveUnknownFields, mapType, listType, listMapKeys, additionalProperties
- **`@status` annotation** - separate status fields or define separate status schema for proper Crossplane resource state management
- **Nested schema expansion** - automatic reference resolution
- **`any` type support** - fields without type constraints for maximum flexibility (IAM policies, etc.)
- **`{any:any}` syntax** - arbitrary property objects with `@preserveUnknownFields`
- **Claims support** - automatic X-prefix handling for composite resources with unprefixed `__xrd_kind`

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

When using `--with-claims`, the tool automatically handles X-prefix naming. With the improvements in this version, you can now use unprefixed `__xrd_kind` for a more natural workflow:

**Using Unprefixed `__xrd_kind` (Recommended):**
```kcl
__xrd_kind = "Bucket"  # No X prefix needed
__xrd_group = "storage.example.org"

# @xrd
schema Bucket:
    name: str
```

```bash
kcl2xrd -i bucket.k --with-claims -o output.yaml

# Generates:
# - XRD Kind: XBucket (X-prefix automatically added)
# - XRD Plural: xbuckets
# - XRD Name: xbuckets.storage.example.org
# - Claim Kind: Bucket (unprefixed)
# - Claim Plural: buckets
```

**Backward Compatible with Prefixed Names:**
```kcl
__xrd_kind = "XDatabase"  # Already has X prefix

# @xrd
schema Database:
    name: str
```

```bash
kcl2xrd -i database.k --with-claims -o output.yaml

# Generates:
# - XRD Kind: XDatabase (keeps X-prefix)
# - Claim Kind: Database (X-prefix removed)
```

## Type Mappings

| KCL Type | OpenAPI Type | CEL Type | Example |
|----------|---------------|----------|---------|
| `str` | `string` | `string` | `name: str` |
| `int` | `integer` | `int` | `count: int` |
| `float` | `number` | `double` | `price: float` |
| `bool` | `boolean` | `bool` | `enabled: bool` |
| `any` | (no type) + `x-kubernetes-preserve-unknown-fields` | `dynamic` | `principal?: any` |
| `[T]` | `array` with `items` | `list` | `tags: [str]` |
| `{K:V}` | `object` with `additionalProperties` | `map` | `labels: {str:str}` |
| `{any:any}` | `object` with `additionalProperties: {}` | `map` | `config: {any:any}` |

**Map Types:** KCL map types like `{str:str}`, `{str:int}`, etc. are converted to OpenAPI `object` type with `additionalProperties` schema. The `additionalProperties` field specifies the type of the map values:
- `{str:str}` → `type: object` with `additionalProperties: { type: string }`
- `{str:int}` → `type: object` with `additionalProperties: { type: integer }`
- `{any:any}` → `type: object` with `additionalProperties: {}` (allows any value type)

This mapping ensures that CEL validations can properly recognize these fields as `map` types, enabling CEL expressions like `self.labels.size()` or `self.config['key']`.

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

### Array Validation Annotations

#### `@minItems(n)`
Sets minimum number of items in arrays.

```kcl
# @minItems(1)
tags: [str]
```

#### `@maxItems(n)`
Sets maximum number of items in arrays.

```kcl
# @maxItems(10)
tags: [str]
```

### String Format Annotations

#### `@format(format)`
Specifies the format for string fields. Common formats include `"date-time"`, `"email"`, `"uuid"`, `"uri"`, `"ipv4"`, `"ipv6"`, etc.

```kcl
# @format("date-time")
createdAt: str

# @format("email")
email: str

# @format("uuid")
id: str
```

### Enum Validation

#### `@enum([values])`
Restricts field to specific allowed values.

```kcl
# @enum(["active", "inactive", "pending"])
status: str
```

### Schema Composition Annotations

#### `@oneOf([[fields]])`
Specifies that exactly one of the given field combinations must be present. This is useful for mutually exclusive options.

```kcl
schema AccessControl:
    groupName?: str
    groupRef?: str
    
    # Exactly one of groupName or groupRef must be provided
    # @oneOf([["groupName"], ["groupRef"]])
    config: {str:str}
```

This generates:
```yaml
oneOf:
  - required: ["groupName"]
  - required: ["groupRef"]
```

#### `@anyOf([[fields]])`
Specifies that at least one of the given field combinations must be present. This is useful for requiring at least one way to identify or configure something.

```kcl
schema User:
    userEmail?: str
    userObjectId?: str
    
    # At least one of userEmail or userObjectId must be provided
    # @anyOf([["userEmail"], ["userObjectId"]])
    userConfig: {str:str}
```

This generates:
```yaml
anyOf:
  - required: ["userEmail"]
  - required: ["userObjectId"]
```

#### Combined `@oneOf` and `@anyOf`
You can use both annotations together for complex validation requirements:

```kcl
schema AccessControl:
    groupName?: str
    groupRef?: str
    userEmail?: str
    userObjectId?: str
    
    # Exactly one group identifier AND at least one user identifier
    # @oneOf([["groupName"], ["groupRef"]])
    # @anyOf([["userEmail"], ["userObjectId"]])
    config: {str:str}
```

This generates both `oneOf` and `anyOf` constraints in the OpenAPI schema, ensuring proper validation of the resource configuration.

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

#### `@status`
Marks a field as a status field, placing it in the `status` section of the XRD instead of `spec.parameters`. Status fields represent the observed state of the resource rather than the desired state.

**Option 1: Status fields in main schema**

```kcl
schema Database:
    # Spec fields (desired state)
    name: str
    size: str
    
    # Status fields (observed state)
    # @status
    ready: bool
    
    # @status
    # @preserveUnknownFields
    conditions?: {any:any}
    
    # @status
    endpoint?: str
```

**Option 2: Separate status schema (recommended)**

You can define status as a separate schema marked with `@status`:

```kcl
# @xrd
schema Application:
    name: str
    replicas: int

# @status
schema AppStatus:
    ready: bool
    phase?: str
    endpoint?: str
```

This generates an XRD with separate `spec` and `status` sections:
- Spec fields go to `spec.parameters`
- Status fields go to `status` (sibling to `spec`)
- All validation and Kubernetes annotations work with status fields

**Empty Status with Preserve Unknown Fields:**

You can also define status without any fields using the `__xrd_status_preserve_unknown_fields` metadata variable:

```kcl
__xrd_group = "example.org"
__xrd_kind = "KafkaCluster"
__xrd_status_preserve_unknown_fields = True

schema KafkaCluster:
    tenant: str
    replicas?: int = 3
```

This generates a status section with just `x-kubernetes-preserve-unknown-fields: true`, allowing any status fields to be set dynamically.

#### `@additionalProperties`
Allows a field to accept additional properties beyond those defined in its schema. Sets `additionalProperties: true` on the field.

```kcl
schema Config:
    # Accept any additional string properties
    # @additionalProperties
    settings: {str:str}
    
    # @status
    # @additionalProperties
    metrics?: {str:int}
```

This generates:
```yaml
settings:
  type: object
  additionalProperties: true
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
    
    # String with format validation
    # @format("date-time")
    createdAt: str
    
    # Email with format and pattern
    # @format("email")
    email?: str
    
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
    
    # Array with minimum and maximum items
    # @minItems(1)
    # @maxItems(10)
    requiredItems: [str]
    
    # CEL validation
    # @validate("self > 0", "Must be positive")
    value: int
    
    # oneOf/anyOf example for mutually exclusive options
    groupName?: str
    groupRef?: str
    userEmail?: str
    userObjectId?: str
    
    # Configuration requiring exactly one group identifier and at least one user identifier
    # @oneOf([["groupName"], ["groupRef"]])
    # @anyOf([["userEmail"], ["userObjectId"]])
    accessConfig?: {str:str}
```

## Metadata Variables

Define in your KCL file with `__xrd_` prefix:

- `__xrd_kind` - Schema kind name (can be variable reference like `_myKind`)
- `__xrd_group` - API group (supports literals, format expressions, property access, and variable references)
- `__xrd_version` - API version (default: v1alpha1)
- `__xrd_categories` - Categories list
- `__xrd_served` - Served flag (True/False)
- `__xrd_referenceable` - Referenceable flag (True/False)
- `__xrd_status_preserve_unknown_fields` - Enable empty status with preserve-unknown-fields (True/False)
- `__xrd_printer_columns` - Printer columns list

### Metadata Variable Resolution with KCL Runtime

All metadata variables are evaluated using the KCL runtime, which provides maximum flexibility:

**1. String Literals:**
```kcl
__xrd_kind = "Database"
__xrd_group = "platform.example.com"
```

**2. Variable References:**
```kcl
_myKind = "PostgreSQL"
_myGroup = "database.example.com"

__xrd_kind = _myKind
__xrd_group = _myGroup
```

**3. Format Expressions:**
```kcl
_subgroup = "storage"
_domain = "mycorp.io"

__xrd_group = "{}.{}".format(_subgroup, _domain)
# Resolves to: storage.mycorp.io
```

**4. Property Access:**
```kcl
settings = {
    PLATFORM_API_GROUP: "platform.example.com"
}

__xrd_group = "{}.{}".format("database", settings.PLATFORM_API_GROUP)
# Resolves to: database.platform.example.com
```

**5. Imports from Central Configuration:**
```kcl
import base.settings

__xrd_group = "{}.{}".format("storage", settings.PLATFORM_API_GROUP)
__xrd_version = settings.DEFAULT_VERSION
# Values resolved from imported settings module
```

### Import Support for Reusable Configuration

You can create central configuration files and import them across multiple XRDs:

**Central Settings File (`base/settings.k`):**
```kcl
# Organization-wide settings
PLATFORM_API_GROUP = "platform.mycorp.io"
DEFAULT_VERSION = "v1alpha1"
DEFAULT_CATEGORIES = ["platform", "managed"]
```

**XRD File Using Imports:**
```kcl
import base.settings

_subgroup = "database"

__xrd_kind = "PostgreSQL"
__xrd_group = "{}.{}".format(_subgroup, settings.PLATFORM_API_GROUP)
__xrd_version = settings.DEFAULT_VERSION

# @xrd
schema PostgreSQL:
    name: str
    size?: int
```

**Result:**
- Group automatically resolves to: `database.platform.mycorp.io`
- Version uses: `v1alpha1` (from central settings)
- No CLI flags needed!

**How it works:**
1. First, the tool tries to evaluate the file with imports intact
2. If imports can be resolved (local modules exist), they work correctly
3. If imports fail (missing dependencies), it falls back to filtering imports and evaluating variable references
4. This ensures maximum flexibility while maintaining robustness

**Note on `__xrd_group`:** 
- Supports any valid KCL expression: string literals, format expressions, property access, variable references
- Automatically evaluated using the KCL runtime for maximum flexibility
- Works with imports from central configuration files
- Examples:
  - `__xrd_group = "example.org"` (literal)
  - `__xrd_group = _myGroup` (variable reference)
  - `__xrd_group = "{}.{}".format(_sub, _domain)` (format expression)
  - `__xrd_group = "{}.{}".format("db", settings.API_GROUP)` (property access)
  - `__xrd_group = "{}.{}".format("db", settings.PLATFORM_API_GROUP)` (imported module)

**Note on `__xrd_kind`:**
- Specifies the `spec.names.kind` for the XRD (e.g., `"Bucket"`, `"Database"`)
- Supports string literals and variable references (e.g., `__xrd_kind = _myKind`)
- When using `--with-claims` flag, accepts unprefixed names:
  - If `__xrd_kind = "Bucket"`, generates XRD kind `XBucket` and claim kind `Bucket`
  - Automatically adds X prefix for XRD, uses unprefixed for claims
  - Backward compatible: if you provide `"XBucket"`, it strips and re-adds X correctly
- The XRD `metadata.name` uses the plural of this kind (e.g., `"buckets.group"` or `"xbuckets.group"` with claims)
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

### Using Status Fields

For Crossplane composite resources, separate desired state (spec) from observed state (status) using the `@status` annotation:

```kcl
schema Database:
    """A database composite resource"""
    
    # Spec fields: what the user wants
    # @pattern("^[a-z0-9-]+$")
    name: str
    
    # @enum(["small", "medium", "large"])
    size?: str = "small"
    
    # @minimum(1)
    replicas?: int = 1
    
    # Status fields: what the system observes
    # @status
    ready: bool
    
    # @status
    phase?: str
    
    # @status
    # @preserveUnknownFields
    conditions?: {any:any}
    
    # @status
    endpoint?: str
```

This generates an XRD with:
- `spec.parameters` containing desired state fields (name, size, replicas)
- `status` section containing observed state fields (ready, phase, conditions, endpoint)

Benefits:
- Clear separation of concerns
- Follows Kubernetes resource conventions
- Status fields can use all annotations (validation, preserveUnknownFields, etc.)
- Required fields are tracked separately for spec and status

### Organizing Configuration with Imports

For teams managing multiple XRDs, create central configuration files to maintain consistency:

**1. Create a central settings file:**

```kcl
# base/settings.k
PLATFORM_API_GROUP = "platform.mycorp.io"
DEFAULT_VERSION = "v1alpha1"
COMMON_CATEGORIES = ["platform", "managed"]
```

**2. Import and use in your XRDs:**

```kcl
# s3bucket.k
import base.settings

__xrd_kind = "S3Bucket"
__xrd_group = "{}.{}".format("storage", settings.PLATFORM_API_GROUP)
__xrd_version = settings.DEFAULT_VERSION

# @xrd
schema S3Bucket:
    name: str
    versioned?: bool
```

```kcl
# database.k
import base.settings

__xrd_kind = "PostgreSQL"
__xrd_group = "{}.{}".format("database", settings.PLATFORM_API_GROUP)
__xrd_version = settings.DEFAULT_VERSION

# @xrd
schema PostgreSQL:
    name: str
    size: int
```

**Benefits:**
- ✅ Consistent API group across all XRDs
- ✅ Easy updates (change once, applies everywhere)
- ✅ Reduced duplication
- ✅ Type-safe with KCL validation

**3. Use variable references for dynamic configuration:**

```kcl
import base.settings

# Define per-resource configuration
_subgroup = "storage"
_resourceType = "S3Bucket"

# Use in metadata
__xrd_kind = _resourceType
__xrd_group = "{}.{}".format(_subgroup, settings.PLATFORM_API_GROUP)

# @xrd
schema S3Bucket:
    name: str
```

This pattern makes it easy to generate multiple related XRDs with consistent naming conventions.

## Examples

See [`examples/`](examples/) directory:

1. **postgresql.k** - Basic schema with optional fields
2. **validated.k** - Validation annotations
3. **nested-schema.k** - Nested schema references
4. **dynatrace-with-metadata.k** - Full in-file metadata
5. **preserve-unknown-fields.k** - Arbitrary properties with `{any:any}`
6. **s3-bucket-with-policy.k** - Complex example with `any` type fields and IAM policies
7. **oneof-anyof-example.k** - Schema composition with oneOf and anyOf validations

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
