"""
Nested Schema Example

This demonstrates nested schemas where one schema references another.
"""

schema MyBucketParams:
    r"""
    Bucket configuration parameters
    
    Attributes
    ----------
    labels : {str:str}, optional
        Labels to apply to the bucket
    region : str, optional
        AWS region for the bucket
    """
    
    labels?: {str:str}
    
    region?: str = "eu-central-1"

schema MyBucket:
    r"""
    S3 Bucket composite resource
    
    Attributes
    ----------
    parameters : MyBucketParams, required
        Configuration parameters for the bucket
    bucketName : str, required
        The name of the S3 bucket
    """
    
    parameters: MyBucketParams
    
    bucketName: str
