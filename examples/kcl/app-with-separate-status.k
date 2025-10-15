"""
Application with Separate Status Schema

This example demonstrates:
1. Separate status schema marked with @status annotation
2. @additionalProperties annotation on individual fields
"""

__xrd_group = "example.org"
__xrd_kind = "Application"
__xrd_version = "v1alpha1"

# @xrd
schema Application:
    """
    An application composite resource
    """
    
    # Spec fields
    # @pattern("^[a-z0-9-]+$")
    name: str
    
    # @minimum(1)
    # @maximum(100)
    replicas?: int = 3
    
    # Config with additionalProperties
    # @additionalProperties
    config?: {str:str}

# @status
schema AppStatus:
    """
    Status for the application
    """
    
    ready: bool
    
    phase?: str
    
    endpoint?: str
    
    # @additionalProperties
    metrics?: {str:int}
