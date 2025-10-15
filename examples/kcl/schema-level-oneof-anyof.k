__xrd_kind = "UserAccount"
__xrd_group = "identity.example.org"
__xrd_version = "v1"
__xrd_served = True
__xrd_referenceable = True

# Schema-level oneOf and anyOf applied to parameters
# @xrd
# @oneOf([["emailAddress"], ["userId"]])
# @anyOf([["password"], ["sshKey"]])
schema UserAccount:
    # Email address for the user
    emailAddress?: str
    # User ID number
    userId?: str
    # Password authentication
    password?: str
    # SSH key authentication
    sshKey?: str
    # User's display name
    displayName: str
