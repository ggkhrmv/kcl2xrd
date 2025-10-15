# Example demonstrating granular control of x-kubernetes-preserve-unknown-fields
# This shows how to apply preserve-unknown-fields only to array items, not the array itself

schema MyApp:
    name: str
    
    # Using @preserveUnknownFields on [{any:any}] applies only to items
    # This generates:
    # filter:
    #   type: array
    #   items:
    #     type: object
    #     x-kubernetes-preserve-unknown-fields: true
    # (note: x-kubernetes-preserve-unknown-fields is NOT on the array itself)
    # @preserveUnknownFields
    filter: [{any:any}]
    
    # Using @itemsPreserveUnknownFields explicitly applies only to items
    # This works with any array type
    # @itemsPreserveUnknownFields
    configs: [{str:str}]
    
    # You can also use @preserveUnknownFields on non-array fields normally
    # @preserveUnknownFields
    metadata?: {any:any}
