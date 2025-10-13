"""
Preserve Unknown Fields Example

This schema demonstrates using {any:any} and [{any:any}] to allow
arbitrary properties in objects and arrays, combined with @preserveUnknownFields annotation.
"""

__xrd_kind = "FlexibleConfig"
__xrd_version = "v1alpha1"
__xrd_group = "config.example.io"

# @xrd
schema FlexibleConfig:
    # Resource name
    name: str
    
    # @preserveUnknownFields
    # Configuration object that can accept any arbitrary key-value pairs
    # This allows users to add custom fields not defined in the schema
    config: {any:any}
    
    # @preserveUnknownFields  
    # List of arbitrary objects for flexible data structures
    # Each item can have any properties
    metadata?: [{any:any}]
    
    # Standard map type (not arbitrary)
    # This only accepts string keys and string values
    labels?: {str:str}
    
    # @preserveUnknownFields
    # Optional settings that can be any structure
    settings?: {any:any}
    
    # Array of strings (not arbitrary)
    tags?: [str]
