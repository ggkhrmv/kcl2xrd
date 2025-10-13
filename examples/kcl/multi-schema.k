"""
Test file with multiple schemas and new features
"""

schema ConfigParams:
    # @preserveUnknownFields
    settings?: {str:str}  # Dynamic configuration settings
    timeout?: int = 30  # Timeout in seconds

schema DynatraceAlerting:
    applicationRef: ApplicationRef
    alertingConfig: AlertingConfig

schema ApplicationRef:
    name: str  # Application name
    namespace?: str = "default"  # Kubernetes namespace

schema AlertingConfig:
    scope?: str = "environment"
    schemaId?: str = "builtin:alerting.profile"
    schemaVersion?: str = "8.5"
    value: AlertingConfigValue

schema AlertingConfigValue:
    name: str  # Profile name
    severityRules?: [AlertingConfigValueSeverityRule]

schema AlertingConfigValueSeverityRule:
    # @enum(["AVAILABILITY", "CUSTOM_ALERT", "ERROR", "PERFORMANCE", "RESOURCE_CONTENTION", "RESOURCE_LIMITATION", "SECURITY", "USAGE"])
    severityLevel: str  # Severity level for the rule
    delayInMinutes: int = 0  # Delay before triggering
    # @enum(["ALL_PROBLEMS", "ONLY_OPEN_PROBLEMS", "ONLY_CLOSED_PROBLEMS"])
    tagFilterIncludeMode: str  # Tag filter mode
    # @preserveUnknownFields
    tagFilter?: [str]  # Tags to filter on
