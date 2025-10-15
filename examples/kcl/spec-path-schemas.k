# Example demonstrating @spec.path annotation for custom spec-level schemas
# This shows how to define separate schemas that become objects directly under spec

__xrd_group = "example.org"
__xrd_kind = "Database"
__xrd_version = "v1alpha1"

# @xrd
schema Database:
    # Regular fields go to spec.parameters
    # @pattern("^[a-z0-9-]+$")
    name: str
    # @minimum(1)
    # @maximum(10)
    replicas?: int = 3
    region?: str = "us-east-1"

# @spec.writeConnectionSecretToRef
schema ConnectionSecretRef:
    # Connection secret reference for storing credentials
    name: str
    namespace?: str

# @spec.publishConnectionDetailsTo
schema PublishConnectionDetails:
    # Publish connection details configuration
    name: str
    # @additionalProperties
    metadata?: {str:str}
    configRef?: str

# @status
schema DatabaseStatus:
    # Status for the database
    ready: bool
    phase?: str
    endpoint?: str
