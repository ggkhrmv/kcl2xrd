# kcl2xrd

A tool to convert KCL (KCL Configuration Language) schemas to Crossplane Composite Resource Definitions (XRDs).

## Quick Start

```bash
# Clone the repository
git clone https://github.com/ggkhrmv/kcl2xrd.git
cd kcl2xrd

# Build the tool
make build

# Generate an XRD from a KCL schema with metadata in the file
./bin/kcl2xrd --input examples/kcl/dynatrace-with-metadata.k \
  --output dynatrace-xrd.yaml

# Or use CLI flags to override file metadata
./bin/kcl2xrd --input examples/kcl/postgresql.k \
  --group database.example.org \
  --output postgresql-xrd.yaml

# Generate with claims support
./bin/kcl2xrd --input examples/kcl/xpostgresql.k \
  --group database.example.org \
  --with-claims \
  --output xpostgresql-xrd.yaml
```

## Overview

This project provides a converter that takes KCL schema files and generates Crossplane XRD YAML manifests. It's inspired by:
- [crossplane-tools](https://github.com/crossplane/crossplane-tools) - XRD generator from Go structs
- [kcl-openapi](https://github.com/kcl-lang/kcl-openapi) - CRDs to KCL schema converter

## Features

- **XRD Metadata in KCL Files**: Define schema selection, API group, version, categories, and more directly in KCL
- **Above-Field Comments**: Multi-line field descriptions for better documentation
- Parse KCL schema files and extract type definitions
- Generate valid Crossplane XRD YAML manifests
- Support for various KCL types: `str`, `int`, `float`, `bool`, arrays, and objects
- Handle optional fields, required fields, and default values
- **Comprehensive validation support**: Regex patterns, enums, ranges, CEL expressions
- **Nested schema support**: Automatically expand referenced schemas
- **Kubernetes annotations**: Support for x-kubernetes-* fields
- Customizable API group and version for generated XRDs

## Installation

### From Source

```bash
git clone https://github.com/ggkhrmv/kcl2xrd.git
cd kcl2xrd
go build -o bin/kcl2xrd ./cmd/kcl2xrd
```

The binary will be available at `./bin/kcl2xrd`

## Usage

### Basic Usage

```bash
kcl2xrd --input <kcl-file> --group <api-group> [--version <version>] [--output <output-file>]
```

### XRD Metadata in KCL Files

You can define XRD metadata directly in your KCL file for automation. This eliminates the need to track which schema to use and what flags to pass:

```kcl
# XRD Metadata - defines how the XRD should be generated
xrKind = "DynatraceAlerting"
xrVersion = "v1alpha1"
group = "monitoring.crossplane.io"
categories = ["monitoring", "alerting"]
served = True
referenceable = True

schema DynatraceAlerting:
    # Above-field comments become field descriptions
    # They support multiple lines for better documentation
    name: str
```

With metadata in the file, you only need:
```bash
kcl2xrd --input myschema.k --output myxrd.yaml
```

CLI flags still work and override file metadata when specified.

### Options

- `-i, --input`: Input KCL schema file (required)
- `-g, --group`: API group for the XRD (required unless specified in KCL file)
- `-v, --version`: API version for the XRD (default: `v1alpha1`)
- `-s, --schema`: Name of the schema to convert (defaults to xrKind or last schema in file)
- `-o, --output`: Output XRD file (if not specified, outputs to stdout)
- `--with-claims`: Generate XRD with claimNames (for creating claimable resources)
- `--claim-kind`: Custom kind for the claim (defaults to schema name without 'X' prefix)
- `--claim-plural`: Custom plural for the claim (auto-generated if not specified)
- `--served`: Mark version as served (default: true)
- `--referenceable`: Mark version as referenceable (default: true)
- `--categories`: Categories for the XRD (comma-separated)
- `--printer-columns`: Additional printer columns (format: `name:type:jsonPath:description`)

### Metadata Variables (in KCL file)

- `xrKind = "SchemaName"`: Specifies which schema to convert
- `xrVersion = "v1alpha1"`: API version for the XRD
- `group = "api.example.org"`: API group for the XRD
- `categories = ["cat1", "cat2"]`: Categories for the XRD
- `served = True/False`: Whether version is served
- `referenceable = True/False`: Whether version is referenceable

### Example

Given a KCL schema file `postgresql.k`:

```kcl
schema PostgreSQLInstance:
    r"""
    PostgreSQL Database Instance
    
    Attributes
    ----------
    storageGB : int, required
        The amount of storage in gigabytes
    instanceSize : str, optional
        The size of the database instance
    version : str, optional
        The PostgreSQL version to deploy
    """
    
    storageGB: int
    
    instanceSize?: str
    
    version?: str = "15"
```

Generate an XRD:

```bash
kcl2xrd --input postgresql.k --group database.example.org --output postgresql.yaml
```

This produces:

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
    name: postgresqlinstances.database.example.org
spec:
    group: database.example.org
    names:
        kind: PostgreSQLInstance
        plural: postgresqlinstances
    versions:
        - name: v1alpha1
          served: true
          referenceable: true
          schema:
            openAPIV3Schema:
                type: object
                properties:
                    spec:
                        type: object
                        properties:
                            parameters:
                                type: object
                                properties:
                                    storageGB:
                                        type: integer
                                    instanceSize:
                                        type: string
                                    version:
                                        type: string
                                        default: "15"
                                required:
                                    - storageGB
                        required:
                            - parameters
                required:
                    - spec
```

### Nested Schema Support

The tool supports nested schemas where one schema references another schema as a field type. The referenced schema's properties are automatically expanded inline in the generated XRD.

Example:

```kcl
schema MyBucketParams:
    labels?: {str:str}
    region?: str = "eu-central-1"

schema MyBucket:
    parameters: MyBucketParams
    bucketName: str
```

This generates an XRD where the `parameters` field is expanded to include all properties from `MyBucketParams`:

```yaml
properties:
  spec:
    properties:
      parameters:
        properties:
          bucketName:
            type: string
          parameters:
            type: object
            properties:
              labels:
                type: object
              region:
                type: string
                default: eu-central-1
```

### Generating XRDs with Claims

Crossplane XRDs can define both composite resources and claims. Claims are a more user-friendly way to provision resources. Use the `--with-claims` flag to generate claimable XRDs:

```bash
kcl2xrd --input xpostgresql.k --group database.example.org --with-claims --output xpostgresql.yaml
```

For a schema named `XPostgreSQLInstance`, this will automatically generate:
- Composite resource: `XPostgreSQLInstance` (kind) / `xpostgresqlinstances` (plural)
- Claim: `PostgreSQLInstance` (kind) / `postgresqlinstances` (plural)

The 'X' prefix is automatically removed from the claim name following Crossplane conventions.

You can also specify custom claim names:

```bash
kcl2xrd --input myresource.k --group example.org --with-claims \
  --claim-kind MyCustomClaim --claim-plural mycustomclaims \
  --output myresource.yaml
```

## Supported KCL Types

The converter supports the following KCL type mappings:

| KCL Type | OpenAPI Type | Example |
|----------|---------------|---------|
| `str` | `string` | `name: str` |
| `int` | `integer` | `count: int` |
| `float` | `number` | `price: float` |
| `bool` | `boolean` | `enabled: bool` |
| `[T]` | `array` with items of type T | `tags: [str]` |
| `{K:V}` | `object` | `labels: {str:str}` |

### Field Modifiers

- **Required fields**: `fieldName: Type`
- **Optional fields**: `fieldName?: Type`
- **Default values**: `fieldName?: Type = value`
- **Field descriptions**: Use above-field comments for multi-line support

Example:
```kcl
schema MyResource:
    # This is the resource name
    # It must be unique within the namespace
    name: str
    
    # Optional field with default value
    region?: str = "us-east-1"
```

## Multi-Schema Files

When a KCL file contains multiple schemas, use the `--schema` flag to specify which one to convert:

```bash
kcl2xrd -i file.k -g api.example.org --schema DynatraceAlerting
```

If not specified, the last schema in the file is used.

## Validation Annotations

The tool supports rich validation annotations using comments to generate OpenAPI validation constraints and Kubernetes-specific validations:

### Basic Validations

```kcl
schema ValidatedResource:
    # @pattern("^[a-z0-9-]+$")
    # @minLength(3)
    # @maxLength(63)
    name: str
    
    # @minimum(0)
    # @maximum(100)
    age?: int
    
    # @enum(["active", "inactive", "pending"])
    status?: str = "pending"
    
    # @immutable
    resourceId: str
```

### Supported Validation Annotations

| Annotation | Applies To | Description | Example |
|------------|-----------|-------------|---------|
| `@pattern("regex")` | string | Regex pattern validation | `@pattern("^[a-z]+$")` |
| `@minLength(n)` | string | Minimum string length | `@minLength(3)` |
| `@maxLength(n)` | string | Maximum string length | `@maxLength(255)` |
| `@minimum(n)` | number | Minimum numeric value | `@minimum(0)` |
| `@maximum(n)` | number | Maximum numeric value | `@maximum(100)` |
| `@enum([...])` | any | Enumeration of allowed values | `@enum(["a", "b", "c"])` |
| `@immutable` | any | Field cannot be changed after creation (x-kubernetes-immutable) | `@immutable` |
| `@validate("rule", "msg")` | any | CEL validation expression | `@validate("self > 0", "Must be positive")` |
| `@preserveUnknownFields` | object/array | Allow additional undefined properties (x-kubernetes-preserve-unknown-fields) | `@preserveUnknownFields` |
| `@mapType("type")` | object | Kubernetes map merge strategy - atomic or granular (x-kubernetes-map-type) | `@mapType("atomic")` |
| `@listType("type")` | array | Kubernetes list type - atomic, set, or map (x-kubernetes-list-type) | `@listType("atomic")` |
| `@listMapKeys(["key"])` | array | Keys for map-type lists (x-kubernetes-list-map-keys) | `@listMapKeys(["name"])` |

For more information on Kubernetes CRD validation fields, see the [Kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/).

### CEL Validations

Use `@validate` for complex validation rules using Common Expression Language (CEL):

```kcl
schema DateRange:
    # @pattern("^\\d{4}-\\d{2}-\\d{2}$")
    startDate: str
    
    # @pattern("^\\d{4}-\\d{2}-\\d{2}$")
    # @validate("self >= self.parent.startDate", "End date must be after start date")
    endDate: str

schema ScalableResource:
    # @minimum(1)
    # @validate("self <= self.parent.maxReplicas", "Min must be <= max")
    minReplicas?: int = 1
    
    # @maximum(100)
    # @validate("self >= self.parent.minReplicas", "Max must be >= min")
    maxReplicas?: int = 10
```

Generated XRD includes:

```yaml
x-kubernetes-validations:
  - rule: self >= self.parent.startDate
    message: End date must be after start date
```

## Supported KCL Types

## Examples

Check the `examples/` directory for more examples:

- `examples/kcl/postgresql.k` - PostgreSQL database instance
- `examples/kcl/k8scluster.k` - Kubernetes cluster with autoscaling
- `examples/kcl/xpostgresql.k` - Composite resource with claims
- `examples/kcl/validated.k` - Schema with validation annotations
- `examples/kcl/advanced-validated.k` - Schema with CEL validation rules
- `examples/kcl/nested-schema.k` - Nested schema references
- `examples/kcl/multi-schema.k` - Multiple schemas with schema selection
- `examples/kcl/dynatrace-with-metadata.k` - XRD metadata defined in KCL file
- `examples/xrd/` - Generated XRD outputs

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/kcl2xrd ./cmd/kcl2xrd
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## How This Fits in the Crossplane Ecosystem

This tool bridges the gap between KCL schemas and Crossplane XRDs, complementing existing tools:

```
┌─────────────────┐
│   Go Structs    │
└────────┬────────┘
         │ crossplane-tools
         ▼
    ┌────────┐
    │  XRDs  │ ◄─── kcl2xrd (this tool)
    └────┬───┘
         │                    ┌──────────────┐
         │                    │ KCL Schemas  │
         │                    └──────▲───────┘
         │                           │
         │                           │ kcl-openapi
         │                           │
         │                    ┌──────┴───────┐
         └───────────────────►│     CRDs     │
                              └──────────────┘
```

**Workflow:**
1. **CRDs → KCL**: Use `kcl-openapi` to convert Kubernetes CRDs to KCL schemas
2. **KCL → XRDs**: Use `kcl2xrd` (this tool) to generate Crossplane XRDs from KCL schemas
3. **Go → XRDs**: Use `crossplane-tools` to generate XRDs from Go structs

This enables teams already using KCL for configuration management to easily create Crossplane composite resources.

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.
