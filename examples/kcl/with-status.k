"""
Database Resource with Status Fields

This example demonstrates how to define status fields using the @status annotation.
Status fields are placed in the status section of the XRD instead of spec.parameters.
"""

__xrd_group = "example.org"
__xrd_kind = "Database"
__xrd_version = "v1alpha1"

schema Database:
    """
    A database composite resource with separate spec and status sections
    """
    
    # Spec fields (desired state)
    # @pattern("^[a-z0-9-]+$")
    name: str
    
    # @enum(["small", "medium", "large"])
    size?: str = "small"
    
    # @minimum(1)
    # @maximum(100)
    replicas?: int = 3
    
    # Status fields (observed state)
    # @status
    ready: bool
    
    # @status
    phase?: str
    
    # @status
    # @preserveUnknownFields
    conditions?: {any:any}
    
    # @status
    endpoint?: str
