# XRD Metadata - using unique __xrd_ prefix to avoid naming conflicts
__xrd_kind = "DynatraceAlerting"
__xrd_version = "v1alpha1"
__xrd_group = "monitoring.crossplane.io"
__xrd_categories = ["monitoring", "alerting", "observability"]
__xrd_served = True
__xrd_referenceable = True
__xrd_printer_columns = ["Status:string:.status.conditions[?(@.type=='Ready')].status:Ready Status", "Age:date:.metadata.creationTimestamp:Age"]

# Helper schemas (not marked with @xrd, so won't be converted on their own)
schema ApplicationRef:
    # The name of the application
    name: str
    
    # The Kubernetes namespace where the application is deployed
    # Defaults to the default namespace
    namespace?: str = "default"

schema AlertingConfig:
    # The scope of the alerting configuration
    scope?: str = "environment"
    
    # The schema identifier for Dynatrace
    schemaId?: str = "builtin:alerting.profile"
    
    # The schema version to use
    schemaVersion?: str = "8.5"
    
    # The alerting configuration value
    value: AlertingConfigValue

schema AlertingConfigValue:
    # The name of the alerting profile
    name: str
    
    # List of severity rules for this alerting profile
    # @listType("atomic")
    severityRules?: [AlertingConfigValueSeverityRule]

schema AlertingConfigValueSeverityRule:
    # The severity level for this rule
    # @enum(["AVAILABILITY", "CUSTOM_ALERT", "ERROR", "PERFORMANCE", "RESOURCE_CONTENTION", "RESOURCE_LIMITATION", "SECURITY", "USAGE"])
    severityLevel: str
    
    # Delay in minutes before triggering the alert
    delayInMinutes: int = 0
    
    # Tag filter include mode for this rule
    # @enum(["ALL_PROBLEMS", "ONLY_OPEN_PROBLEMS", "ONLY_CLOSED_PROBLEMS"])
    tagFilterIncludeMode: str
    
    # List of tags to filter on
    # Optional field that can be used to filter alerts
    tagFilter?: [str]

# Main schema marked with @xrd annotation - this is the parent schema
# @xrd
schema DynatraceAlerting:
    # Reference to the application this alerting configuration applies to
    applicationRef: ApplicationRef
    
    # The alerting configuration details
    alertingConfig: AlertingConfig

# Other schemas that should NOT be converted (not referenced by the @xrd schema)
schema SomeOtherConfig:
    # This schema won't be converted because it's not referenced
    # by the DynatraceAlerting schema (the @xrd marked parent)
    someField: str
    
# Additional KCL code that's not part of XRD definition
# myVariable = "this is just KCL code"
# myFunction = lambda x: x + 1
