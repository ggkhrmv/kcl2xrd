"""
Advanced Validated Resource Schema

This schema demonstrates CEL validation rules in addition to standard validations.
"""

schema AdvancedValidatedResource:
    r"""
    A resource with CEL validation rules
    
    Attributes
    ----------
    startDate : str, required
        Start date in ISO format
    endDate : str, required
        End date must be after start date (validated with CEL)
    replicas : int, optional
        Number of replicas
    minReplicas : int, optional
        Minimum replicas for autoscaling
    maxReplicas : int, optional
        Maximum replicas for autoscaling
    """
    
    # @pattern("^\\d{4}-\\d{2}-\\d{2}$")
    startDate: str
    
    # @pattern("^\\d{4}-\\d{2}-\\d{2}$")
    # @validate("self >= self.parent.startDate", "End date must be after or equal to start date")
    endDate: str
    
    # @minimum(1)
    # @maximum(100)
    replicas?: int = 3
    
    # @minimum(1)
    # @validate("self <= self.parent.maxReplicas", "Min replicas must be less than or equal to max replicas")
    minReplicas?: int = 1
    
    # @maximum(100)
    # @validate("self >= self.parent.minReplicas", "Max replicas must be greater than or equal to min replicas")
    maxReplicas?: int = 10
