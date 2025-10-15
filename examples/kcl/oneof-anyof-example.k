__xrd_kind = "AccessControl"
__xrd_group = "security.example.org"
__xrd_version = "v1"
__xrd_served = True
__xrd_referenceable = True

# @xrd
schema AccessControl:
    # Group identification - exactly one required
    groupName?: str
    groupRef?: str
    # User identification - at least one required
    userEmail?: str
    userObjectId?: str
    # Permission level
    # @enum(["read", "write", "admin"])
    permission: str = "read"
    # Configuration with oneOf and anyOf constraints
    # @oneOf([["groupName"], ["groupRef"]])
    # @anyOf([["userEmail"], ["userObjectId"]])
    config: {str:str}
