"""
Example demonstrating KCL map types mapping to OpenAPI additionalProperties
which enables CEL map validations.
"""

# @xrd
schema ConfigurationMap:
    # @pattern("^[a-z0-9-]+$")
    # @minLength(3)
    # @maxLength(63)
    # Name of the configuration
    name: str
    
    # String to string map - useful for labels, annotations, tags
    # Maps to: type: object with additionalProperties: { type: string }
    # CEL recognizes this as map<string, string>
    # @validate("size(self) <= 10", "Maximum 10 labels allowed")
    labels: {str:str}
    
    # String to int map - useful for resource limits, counts
    # Maps to: type: object with additionalProperties: { type: integer }
    # CEL recognizes this as map<string, int>
    # @validate("self.all(k, self[k] >= 0)", "All counts must be non-negative")
    resourceCounts?: {str:int}
    
    # String to bool map - useful for feature flags
    # Maps to: type: object with additionalProperties: { type: boolean }
    # CEL recognizes this as map<string, bool>
    featureFlags?: {str:bool}
    
    # String to float map - useful for metrics, weights
    # Maps to: type: object with additionalProperties: { type: number }
    # CEL recognizes this as map<string, double>
    # @validate("self.all(k, self[k] >= 0.0 && self[k] <= 1.0)", "Weights must be between 0 and 1")
    weights?: {str:float}
    
    # Any to any map - for completely arbitrary key-value data
    # Maps to: type: object with additionalProperties: {}
    # CEL recognizes this as map<dynamic, dynamic>
    # @preserveUnknownFields
    metadata?: {any:any}
