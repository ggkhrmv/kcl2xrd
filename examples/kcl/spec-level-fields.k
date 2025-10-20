# Example demonstrating @spec annotation for spec-level fields
# This shows how to define fields that go directly under spec, not spec.parameters

schema MyComposite:
    # Regular fields - these go to spec.parameters
    name: str
    
    # @minimum(1)
    replicas?: int = 3
    
    region?: str = "us-east-1"
    
    # Spec-level fields - these go directly under spec (not spec.parameters)
    # This is useful for Crossplane composition-related fields
    
    # @spec
    compositionSelector?: {str:str}
    
    # @spec
    compositionRef?: str
    
    # @spec
    # @immutable
    compositionRevisionRef?: str
    
    # Status fields - these go to status section
    # @status
    ready?: bool
    
    # @status
    phase?: str
