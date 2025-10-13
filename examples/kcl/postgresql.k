"""
PostgreSQL Database Instance Schema

This schema defines a PostgreSQL database instance with configurable storage and instance size.
"""

schema PostgreSQLInstance:
    r"""
    PostgreSQL Database Instance
    
    A composite resource that represents a PostgreSQL database instance
    with configurable parameters for storage, size, and networking.
    
    Attributes
    ----------
    storageGB : int, required
        The amount of storage in gigabytes allocated to the database instance
    instanceSize : str, optional
        The size of the database instance (e.g., small, medium, large)
    version : str, optional
        The PostgreSQL version to deploy (e.g., "13", "14", "15")
    backupRetentionDays : int, optional
        Number of days to retain automated backups
    """
    
    storageGB: int
    
    instanceSize?: str
    
    version?: str = "15"
    
    backupRetentionDays?: int = 7
