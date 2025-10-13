"""
Kubernetes Cluster Schema

This schema defines a Kubernetes cluster with various configuration options.
"""

schema KubernetesCluster:
    r"""
    Kubernetes Cluster composite resource
    
    Attributes
    ----------
    nodeCount : int, required
        The number of worker nodes in the cluster
    nodeSize : str, required
        The size/type of worker nodes
    region : str, required
        The cloud region where the cluster will be deployed
    version : str, optional
        The Kubernetes version
    enableAutoScaling : bool, optional
        Whether to enable autoscaling for the node pool
    minNodes : int, optional
        Minimum number of nodes when autoscaling is enabled
    maxNodes : int, optional
        Maximum number of nodes when autoscaling is enabled
    tags : [str], optional
        List of tags to apply to cluster resources
    highAvailability : bool, optional
        Whether to deploy the cluster in high availability mode
    """
    
    nodeCount: int
    
    nodeSize: str
    
    region: str
    
    version?: str = "1.28"
    
    enableAutoScaling?: bool = False
    
    minNodes?: int = 1
    
    maxNodes?: int = 10
    
    tags?: [str]
    
    highAvailability?: bool = False
