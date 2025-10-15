# Example demonstrating $ prefix for internal/private fields
# Fields with $ prefix are KCL internal variables and are omitted from XRD generation

schema MyApp:
    # Regular fields - these will be in the XRD
    name: str
    # @minimum(1)
    replicas?: int = 3
    
    # Internal KCL variables - these will be omitted from XRD
    # $filter is a KCL internal keyword/variable
    $filter: str
    
    # Another internal variable
    $internalConfig: {str:str}
    
    # Regular field - will be in the XRD
    region?: str = "us-east-1"
