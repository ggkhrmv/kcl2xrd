"""
Composite PostgreSQL Instance Schema

This schema defines a composite PostgreSQL database instance (XPostgreSQLInstance)
that can be claimed as a PostgreSQLInstance.
"""

schema XPostgreSQLInstance:
    r"""
    Composite PostgreSQL Database Instance
    
    A composite resource that represents a PostgreSQL database instance.
    This can be claimed using the PostgreSQLInstance claim.
    
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
    highAvailability : bool, optional
        Whether to enable high availability mode
    """
    
    storageGB: int
    
    instanceSize?: str = "medium"
    
    version?: str = "15"
    
    backupRetentionDays?: int = 7
    
    highAvailability?: bool = False
