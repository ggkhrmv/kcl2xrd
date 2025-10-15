__xrd_kind = "ContactList"
__xrd_group = "directory.example.org"
__xrd_version = "v1"
__xrd_served = True
__xrd_referenceable = True

# @xrd
schema ContactList:
    # List of email addresses with email format validation
    # @itemsFormat("email")
    emails: [str]
    # List of phone numbers with specific format
    # @itemsFormat("phone")
    # @minItems(1)
    phoneNumbers?: [str]
    # List of website URLs
    # @itemsFormat("uri")
    websites?: [str]
    # List of unique identifiers
    # @itemsFormat("uuid")
    identifiers?: [str]
