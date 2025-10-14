"""
S3 Bucket with IAM Policy Example

This example demonstrates:
- Using @xrd annotation to mark the main schema
- Using 'any' type for flexible IAM policy fields
- Module-level variables that won't be parsed as schema fields
- Nested schema references
"""

# XRD Metadata
__xrd_kind = "Bucket"
__xrd_group = "aws.example.org"
__xrd_version = "v1alpha1"
__xrd_categories = ["aws", "storage", "s3"]
__xrd_served = True
__xrd_referenceable = True

# Nested schemas for bucket policy structure
schema BucketPolicyStatement:
    # @enum(["Allow", "Deny"])
    # Effect of the policy statement (Allow/Deny)
    effect?: str
    
    # @preserveUnknownFields
    # The principals this statement applies to
    # Can be a string, array, or object
    principal?: any
    
    # @preserveUnknownFields
    # Action(s) that the policy allows/denies
    # Can be a string or array
    action?: any
    
    # @preserveUnknownFields
    # Resource(s) to which this policy applies
    # Can be a string or array
    resource?: any
    
    # @preserveUnknownFields
    # Conditions for when the policy is in effect
    condition?: any

schema BucketPolicy:
    # Policy version defaults to 2012-10-17 if not specified
    version?: str = "2012-10-17"
    
    # List of policy statements
    statement: [BucketPolicyStatement]

# Main schema - marked with @xrd
# @xrd
schema Bucket:
    # @pattern("^[a-z0-9][a-z0-9-]*[a-z0-9]$")
    # @minLength(3)
    # @maxLength(63)
    # The name of the S3 Bucket
    name: str
    
    # Whether versioning is enabled for the bucket
    versioned?: bool = False
    
    # @preserveUnknownFields
    # The bucket policy in structured format
    policy?: BucketPolicy
    
    # @enum(["us-east-1", "us-west-2", "eu-west-1", "eu-central-1"])
    # AWS region for the bucket
    region?: str = "us-east-1"

# Module-level variables (won't be parsed as schema fields)
_composition_template = {
    "kind": "Composition"
    "metadata": {
        "name": "s3bucket"
    }
}
