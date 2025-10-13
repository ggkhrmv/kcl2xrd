"""
Advanced Application Configuration Schema

This schema demonstrates nested objects and complex types.
"""

schema ApplicationConfig:
    r"""
    Application Configuration with nested settings
    
    Attributes
    ----------
    name : str, required
        The name of the application
    replicas : int, optional
        Number of replicas
    resources : ResourceRequirements, optional
        Resource requirements for the application
    networking : NetworkConfig, optional
        Network configuration
    """
    
    name: str
    
    replicas?: int = 3
    
    resources?: ResourceRequirements
    
    networking?: NetworkConfig

schema ResourceRequirements:
    r"""
    Resource requirements
    
    Attributes
    ----------
    cpu : str, optional
        CPU request/limit
    memory : str, optional
        Memory request/limit
    """
    
    cpu?: str = "100m"
    
    memory?: str = "128Mi"

schema NetworkConfig:
    r"""
    Network configuration
    
    Attributes
    ----------
    port : int, optional
        Service port
    protocol : str, optional
        Network protocol
    """
    
    port?: int = 8080
    
    protocol?: str = "TCP"
