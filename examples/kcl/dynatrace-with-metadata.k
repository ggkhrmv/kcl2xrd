# XRD Metadata - defines how the XRD should be generated
xrKind = "DynatraceAlerting"
xrVersion = "v1alpha1"
group = "monitoring.crossplane.io"
categories = ["monitoring", "alerting"]
served = True
referenceable = True

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

schema DynatraceAlerting:
    # Reference to the application this alerting configuration applies to
    applicationRef: ApplicationRef
    
    # The alerting configuration details
    alertingConfig: AlertingConfig
