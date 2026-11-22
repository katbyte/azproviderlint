package AZS016

// products are the names checked from the start, written the way they should be, under the service the provider's documentation files them
// under: every azurerm resource whose name is two plain words or more, and the services and ideas that descriptions name most.
var products = []string{
	// Active Directory Domain Services
	"Active Directory Domain Service", "Active Directory Domain Service Trust",

	// Analysis Services
	"Analysis Services Server",

	// App Configuration
	"App Configuration", "App Configuration Feature", "App Configuration Key",

	// App Service (Web Apps)
	"App Service", "App Service Certificate Binding", "App Service Certificate Order", "App Service Environment", "App Service Plan",
	"App Service Public Certificate", "App Service Source Control Slot", "Deployment Slot", "Function App", "Function App Active Slot",
	"Function App Function", "Function App Hybrid Connection", "Linux Function App", "Linux Function App Slot", "Linux Web App", "Linux Web App Slot",
	"Service Plan", "Static Web App", "Static Web App Custom Domain", "Web App", "Web App Active Slot", "Web App Hybrid Connection",
	"Windows Function App", "Windows Function App Slot", "Windows Web App", "Windows Web App Slot",

	// Application Insights
	"Application Insights", "Application Insights Analytics Item", "Application Insights Smart Detection Rule", "Application Insights Workbook Template",

	// Arc Resource Bridge
	"Arc Resource Bridge Appliance",

	// ArcKubernetes
	"Arc Kubernetes Cluster", "Arc Kubernetes Cluster Extension", "Arc Kubernetes Flux Configuration", "Arc Kubernetes Provisioned Cluster",

	// Attestation
	"Attestation Provider",

	// Authorization
	"Active Directory", "Federated Identity Credential", "Managed Identity", "Managed Service Identity", "Role Assignment", "Role Definition",
	"Service Principal", "System Assigned Identity", "System Assigned Managed Identity", "User Assigned Identity", "User Assigned Managed Identity",

	// Automanage
	"Automanage Configuration",

	// Automation
	"Automation Account", "Automation Certificate", "Automation Connection", "Automation Credential", "Automation Hybrid Runbook Worker",
	"Automation Hybrid Runbook Worker Group", "Automation Module", "Automation Runbook", "Automation Runtime Environment", "Automation Schedule",
	"Automation Source Control",

	// Azure Managed Lustre File System
	"Managed Lustre File System",

	// Base
	"Resource Group", "Resource Manager", "Resource Provider",

	// Batch
	"Batch Account", "Batch Application", "Batch Job",

	// Blueprints
	"Blueprint Assignment",

	// Bot
	"Bot Channels Registration", "Bot Connection", "Bot Web App",

	// Chaos Studio
	"Chaos Studio Capability", "Chaos Studio Experiment", "Chaos Studio Target",

	// Cognitive Services
	"Cognitive Account", "Cognitive Account Project",

	// Communication
	"Communication Service", "Email Communication Service", "Email Communication Service Domain", "Email Communication Service Domain Sender Username",

	// Compute
	"Availability Set", "Availability Zone", "Capacity Reservation", "Capacity Reservation Group", "Dedicated Host", "Disk Access", "Disk Encryption Set",
	"Gallery Application", "Gallery Application Version", "Linux Virtual Machine", "Linux Virtual Machine Scale Set", "Managed Disk",
	"Orchestrated Virtual Machine Scale Set", "Proximity Placement Group", "Shared Image", "Shared Image Gallery", "Virtual Machine",
	"Virtual Machine Extension", "Virtual Machine Gallery Application Assignment", "Virtual Machine Restore Point",
	"Virtual Machine Restore Point Collection", "Virtual Machine Run Command", "Virtual Machine Scale Set", "Windows Virtual Machine",
	"Windows Virtual Machine Scale Set",

	// Confidential Ledger
	"Confidential Ledger",

	// Container
	"Container Connected Registry", "Container Group", "Container Registry Agent Pool", "Container Registry Cache Rule",
	"Container Registry Credential Set", "Container Registry Task", "Container Registry Token Password", "Container Registry Webhook",
	"Kubernetes Cluster Extension", "Kubernetes Fleet Manager", "Kubernetes Fleet Member", "Kubernetes Fleet Update Run",
	"Kubernetes Fleet Update Strategy", "Kubernetes Flux Configuration",

	// Container Apps
	"Container App", "Container App Custom Domain", "Container App Environment", "Container App Environment Certificate",
	"Container App Environment Managed Certificate", "Container App Environment Storage", "Container App Job",

	// Cost Management
	"Cost Anomaly Alert", "Cost Management Scheduled Action",

	// Custom Providers
	"Custom Provider",

	// Dashboard
	"Dashboard Grafana", "Dashboard Grafana Managed Private Endpoint",

	// Data Explorer
	"Kusto Cluster Principal Assignment", "Kusto Script",

	// Data Factory
	"Data Factory", "Data Factory Managed Private Endpoint",

	// Data Share
	"Data Share", "Data Share Account",

	// Databox Edge
	"Databox Edge Device",

	// Databricks
	"Databricks Access Connector", "Databricks Virtual Network Peering", "Databricks Workspace",

	// Desktop Virtualization
	"Virtual Desktop Application", "Virtual Desktop Application Group", "Virtual Desktop Host Pool", "Virtual Desktop Scaling Plan",
	"Virtual Desktop Scaling Plan Host Pool Association", "Virtual Desktop Workspace", "Virtual Desktop Workspace Application Group Association",

	// Dev Center
	"Dev Center", "Dev Center Attached Network", "Dev Center Catalog", "Dev Center Dev Box Definition", "Dev Center Environment Type",
	"Dev Center Gallery", "Dev Center Network Connection", "Dev Center Project", "Dev Center Project Environment Type", "Dev Center Project Pool",

	// Dev Test
	"Dev Test Lab",

	// Digital Twins
	"Digital Twins", "Digital Twins Time Series Database Connection",

	// Fabric
	"Fabric Capacity",

	// Fluid Relay
	"Fluid Relay Server",

	// Graph Services
	"Graph Services Account",

	// Hardware Security Module
	"Dedicated Hardware Security Module",

	// Healthcare
	"Healthcare Med Tech Service", "Healthcare Service",

	// Hybrid Compute
	"Arc Private Link Scope",

	// Key Vault
	"Key Vault", "Key Vault Access Policy", "Key Vault Certificate", "Key Vault Certificate Contacts", "Key Vault Certificate Issuer", "Key Vault Key",
	"Key Vault Managed Hardware Security Module", "Key Vault Managed Hardware Security Module Key",
	"Key Vault Managed Hardware Security Module Role Definition", "Key Vault Managed Storage Account", "Key Vault Secret",

	// Load Test
	"Playwright Workspace",

	// Log Analytics
	"Log Analytics", "Log Analytics Cluster", "Log Analytics Cluster Customer Managed Key", "Log Analytics Data Export Rule",
	"Log Analytics Linked Service", "Log Analytics Linked Storage Account", "Log Analytics Query Pack", "Log Analytics Query Pack Query",
	"Log Analytics Storage Insights", "Log Analytics Workspace",

	// Logic App
	"Logic App", "Logic App Integration Account", "Logic App Integration Account Agreement", "Logic App Integration Account Assembly",
	"Logic App Integration Account Batch Configuration", "Logic App Integration Account Certificate", "Logic App Integration Account Map",
	"Logic App Integration Account Partner", "Logic App Integration Account Schema", "Logic App Integration Account Session", "Logic App Workflow",

	// Machine Learning
	"Machine Learning Compute Cluster", "Machine Learning Compute Instance", "Machine Learning Inference Cluster", "Machine Learning Workspace",
	"Machine Learning Workspace Network Outbound Rule Private Endpoint", "Machine Learning Workspace Network Outbound Rule Service Tag",

	// Managed Applications
	"Managed Application", "Managed Application Definition",

	// Managed Redis
	"Managed Redis Access Policy Assignment",

	// Management
	"Management Group", "Management Group Subscription Association", "Management Lock", "Resource Management Private Link",
	"Resource Management Private Link Association",

	// Maps
	"Maps Account",

	// Messaging
	"Event Grid", "Event Grid Namespace Topic", "Event Grid Partner Configuration", "Event Grid Partner Namespace", "Event Grid System Topic", "Event Hub",
	"Notification Hub", "Notification Hub Namespace", "Relay Hybrid Connection", "Relay Hybrid Connection Authorization Rule", "Relay Namespace",
	"Relay Namespace Authorization Rule", "Service Bus", "Service Bus Namespace Customer Managed Key",

	// Mongo Cluster
	"Mongo Cluster Firewall Rule", "Mongo Cluster User",

	// Monitor
	"Monitor Private Link Scope", "Monitor Private Link Scoped Service", "Monitor Smart Detector Alert Rule", "Monitor Workspace",

	// Network
	"Application Gateway", "Application Security Group", "Bastion Host", "Express Route Circuit Connection", "Express Route Connection",
	"Express Route Port", "Firewall Policy", "Firewall Policy Rule Collection Group", "Front Door", "Local Network Gateway", "Network Connection Monitor",
	"Network Manager", "Network Manager Admin Rule", "Network Manager Admin Rule Collection", "Network Manager Connectivity Configuration",
	"Network Manager Deployment", "Network Manager Management Group Connection", "Network Manager Network Group", "Network Manager Routing Configuration",
	"Network Manager Routing Rule", "Network Manager Routing Rule Collection", "Network Manager Scope Connection",
	"Network Manager Security Admin Configuration", "Network Manager Static Member", "Network Manager Subscription Connection",
	"Network Manager Verifier Workspace", "Network Manager Verifier Workspace Reachability Analysis Intent", "Network Profile", "Network Security Group",
	"Network Security Perimeter", "Network Security Perimeter Access Rule", "Network Security Perimeter Association", "Network Security Perimeter Profile",
	"Network Security Rule", "Network Watcher", "Network Watcher Flow Log", "Private Endpoint", "Private Link", "Private Link Service", "Route Filter",
	"Route Map", "Route Server", "Route Table", "Subnet Service Endpoint Storage Policy", "Traffic Manager", "Traffic Manager Profile", "Virtual Hub",
	"Virtual Hub Route Table", "Virtual Hub Routing Intent", "Virtual Network", "Virtual Network Gateway", "Web Application Firewall Policy",

	// Network Function
	"Network Function Azure Traffic Collector", "Network Function Collector Policy",

	// Oracle
	"Oracle Autonomous Database Backup", "Oracle Resource Anchor",

	// Palo Alto
	"Palo Alto Local Rulestack Prefix List", "Palo Alto Local Rulestack Rule", "Palo Alto Next Generation Firewall Virtual Hub Strata Cloud Manager",
	"Palo Alto Next Generation Firewall Virtual Network Panorama", "Palo Alto Next Generation Firewall Virtual Network Strata Cloud Manager",

	// Policy
	"Management Group Policy Exemption", "Management Group Policy Remediation", "Policy Set Definition", "Resource Group Policy Assignment",
	"Resource Group Policy Exemption", "Resource Group Policy Remediation", "Resource Policy Exemption", "Resource Policy Remediation",
	"Subscription Policy Assignment", "Subscription Policy Exemption", "Subscription Policy Remediation",

	// Portal
	"Portal Tenant Configuration",

	// Purview
	"Purview Account",

	// Recovery Services
	"Backup Protected File Share", "Backup Vault", "Recovery Services Vault", "Site Recovery", "Site Recovery Replication Recovery Plan",

	// Redis
	"Redis Cache", "Redis Cache Access Policy", "Redis Cache Access Policy Assignment", "Redis Linked Server",

	// Search
	"Search Service",

	// Security Center
	"Security Center Assessment", "Security Center Automation",

	// Sentinel
	"Sentinel Automation Rule", "Sentinel Metadata", "Sentinel Threat Intelligence Indicator", "Sentinel Watchlist", "Sentinel Watchlist Item",

	// Service Fabric
	"Service Fabric", "Service Fabric Cluster",

	// Service Networking
	"Application Load Balancer Security Policy",

	// Spring Cloud
	"Spring Cloud", "Spring Cloud Accelerator", "Spring Cloud Application Live View", "Spring Cloud Build Deployment", "Spring Cloud Build Pack Binding",
	"Spring Cloud Builder", "Spring Cloud Certificate", "Spring Cloud Configuration Service", "Spring Cloud Container Deployment",
	"Spring Cloud Custom Domain", "Spring Cloud Customized Accelerator", "Spring Cloud Dev Tool Portal", "Spring Cloud Gateway",
	"Spring Cloud Gateway Custom Domain", "Spring Cloud Gateway Route Config", "Spring Cloud Service", "Spring Cloud Storage",

	// Storage
	"Data Lake", "Storage Account", "Storage Account Local User", "Storage Blob Inventory Policy", "Storage Container", "Storage Encryption Scope",
	"Storage Object Replication", "Storage Queue", "Storage Share", "Storage Sync", "Storage Sync Cloud Endpoint", "Storage Sync Group",
	"Storage Sync Server Endpoint", "Storage Table",

	// Storage Mover
	"Storage Mover", "Storage Mover Agent", "Storage Mover Job Definition", "Storage Mover Project", "Storage Mover Source Endpoint",
	"Storage Mover Target Endpoint",

	// Stream Analytics
	"Stream Analytics Cluster", "Stream Analytics Job", "Stream Analytics Job Schedule", "Stream Analytics Job Storage Account",
	"Stream Analytics Managed Private Endpoint", "Stream Analytics Output Function", "Stream Analytics Output Table",
	"Stream Analytics Reference Input Blob", "Stream Analytics Stream Input Blob",

	// Synapse
	"Synapse Linked Service", "Synapse Managed Private Endpoint", "Synapse Private Link Hub", "Synapse Role Assignment", "Synapse Spark Pool",
	"Synapse Workspace", "Synapse Workspace Extended Auditing Policy",

	// System Center Virtual Machine Manager
	"System Center Virtual Machine Manager Availability Set", "System Center Virtual Machine Manager Cloud",
	"System Center Virtual Machine Manager Server", "System Center Virtual Machine Manager Virtual Machine Instance",
	"System Center Virtual Machine Manager Virtual Machine Instance Guest Agent", "System Center Virtual Machine Manager Virtual Machine Template",
	"System Center Virtual Machine Manager Virtual Network",

	// Template
	"Resource Group Template Deployment", "Subscription Template Deployment",

	// Trusted Signing
	"Trusted Signing Account",

	// Video Indexer
	"Video Indexer Account",
}
