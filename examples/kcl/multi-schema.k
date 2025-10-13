"""
Test file with multiple schemas demonstrating schema selection and annotations
"""

schema MultiSchemaConfigParams:
    # @preserveUnknownFields
    settings?: {str:str}  # Dynamic configuration settings
    timeout?: int = 30  # Timeout in seconds

schema MultiSchemaMonitoring:
    appRef: MultiSchemaAppRef
    config: MultiSchemaAlertConfig

schema MultiSchemaAppRef:
    name: str  # Application name
    namespace?: str = "default"  # Kubernetes namespace

schema MultiSchemaAlertConfig:
    scope?: str = "environment"
    schemaId?: str = "builtin:alerting.profile"
    schemaVersion?: str = "8.5"
    value: MultiSchemaAlertValue

schema MultiSchemaAlertValue:
    name: str  # Profile name
    rules?: [MultiSchemaAlertRule]

schema MultiSchemaAlertRule:
    # @enum(["AVAILABILITY", "CUSTOM_ALERT", "ERROR", "PERFORMANCE", "RESOURCE_CONTENTION", "RESOURCE_LIMITATION", "SECURITY", "USAGE"])
    severityLevel: str  # Severity level for the rule
    delayInMinutes: int = 0  # Delay before triggering
    # @enum(["ALL_PROBLEMS", "ONLY_OPEN_PROBLEMS", "ONLY_CLOSED_PROBLEMS"])
    tagFilterIncludeMode: str  # Tag filter mode
    # @preserveUnknownFields
    tagFilter?: [str]  # Tags to filter on
