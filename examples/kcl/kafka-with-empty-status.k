"""
Kafka Cluster with Empty Status

This example demonstrates:
1. additionalProperties: true on both parameters and status (automatically added)
2. Empty status section with x-kubernetes-preserve-unknown-fields
"""

__xrd_group = "example.org"
__xrd_kind = "KafkaCluster"
__xrd_version = "v1alpha1"
__xrd_status_preserve_unknown_fields = True

schema KafkaCluster:
    """
    A Kafka cluster composite resource
    """
    
    # Spec fields
    # @pattern("^[a-z0-9-]+$")
    tenant: str
    
    # @minimum(1)
    # @maximum(100)
    replicas?: int = 3
    
    # Additional settings
    config?: {str:str}
