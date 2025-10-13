# kcl2xrd

A tool to convert KCL (KCL Configuration Language) schemas to Crossplane Composite Resource Definitions (XRDs).

## Overview

This project provides a converter that takes KCL schema files and generates Crossplane XRD YAML manifests. It's inspired by:
- [crossplane-tools](https://github.com/crossplane/crossplane-tools) - XRD generator from Go structs
- [kcl-openapi](https://github.com/kcl-lang/kcl-openapi) - CRDs to KCL schema converter

## Features

- Parse KCL schema files and extract type definitions
- Generate valid Crossplane XRD YAML manifests
- Support for various KCL types: `str`, `int`, `float`, `bool`, arrays, and objects
- Handle optional fields, required fields, and default values
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

```bash
kcl2xrd --input <kcl-file> --group <api-group> [--version <version>] [--output <output-file>]
```

### Options

- `-i, --input`: Input KCL schema file (required)
- `-g, --group`: API group for the XRD (required)
- `-v, --version`: API version for the XRD (default: `v1alpha1`)
- `-o, --output`: Output XRD file (if not specified, outputs to stdout)

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

## Examples

Check the `examples/` directory for more examples:

- `examples/kcl/postgresql.k` - PostgreSQL database instance
- `examples/kcl/k8scluster.k` - Kubernetes cluster with autoscaling
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

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.
