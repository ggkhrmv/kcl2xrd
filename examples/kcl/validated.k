"""
Validated Resource Schema

This schema demonstrates validation constraints, regex patterns, and immutability.
"""

schema ValidatedResource:
    r"""
    A resource with comprehensive validation rules
    
    Attributes
    ----------
    name : str, required
        The name must match a DNS-compatible pattern
    age : int, optional
        Age must be between 0 and 120
    email : str, optional
        Email must match a valid email pattern
    status : str, optional
        Status must be one of the allowed values
    resourceId : str, required
        Immutable resource identifier
    description : str, optional
        Description with length constraints
    """
    
    # @pattern("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
    # @minLength(3)
    # @maxLength(63)
    name: str
    
    # @minimum(0)
    # @maximum(120)
    age?: int
    
    # @pattern("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$")
    email?: str
    
    # @enum(["pending", "active", "inactive", "deleted"])
    status?: str = "pending"
    
    # @immutable
    # @pattern("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")
    resourceId: str
    
    # @minLength(10)
    # @maxLength(500)
    description?: str
