package client

// Blueprint represents an Iru Blueprint.
type Blueprint struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Icon           string `json:"icon,omitempty"`
	Color          string `json:"color,omitempty"`
	Type           string `json:"type,omitempty"`
	EnrollmentCode struct {
		Code     string `json:"code"`
		IsActive bool   `json:"is_active"`
	} `json:"enrollment_code,omitempty"`
	Source struct {
		ID   string `json:"id,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"source,omitempty"`
}

// ADEIntegration represents an Iru ADE Integration.
type ADEIntegration struct {
	ID                  string     `json:"id,omitempty"`
	BlueprintID         string     `json:"blueprint_id,omitempty"` // Used for Update input
	Phone               string     `json:"phone,omitempty"`        // Top level in some cases?
	Email               string     `json:"email,omitempty"`        // Top level in some cases?
	Blueprint           *Blueprint `json:"blueprint,omitempty"`    // From response
	AccessTokenExpiry   string     `json:"access_token_expiry,omitempty"`
	ServerName          string     `json:"server_name,omitempty"`
	ServerUUID          string     `json:"server_uuid,omitempty"`
	AdminID             string     `json:"admin_id,omitempty"`
	OrgName             string     `json:"org_name,omitempty"`
	STokenFileName      string     `json:"stoken_file_name,omitempty"`
	DaysLeft            int        `json:"days_left,omitempty"`
	Status              string     `json:"status,omitempty"`
	UseBlueprintRouting bool       `json:"use_blueprint_routing,omitempty"`
	Defaults            struct {
		Phone string `json:"phone"`
		Email string `json:"email"`
	} `json:"defaults,omitempty"`
}

// Device represents an Iru Device.
type Device struct {
	ID           string `json:"device_id,omitempty"`
	DeviceName   string `json:"device_name,omitempty"`
	AssetTag     string `json:"asset_tag,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	Model        string `json:"model,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`
	BlueprintID  string `json:"blueprint_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Platform     string `json:"platform,omitempty"`
	LastCheckIn  string `json:"last_check_in,omitempty"`
}

// DeviceNote represents a note assigned to a device.
type DeviceNote struct {
	ID        string `json:"note_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Author    string `json:"author,omitempty"`
	Content   string `json:"content"`
}

// DeviceDetails represents the full details of a device.
type DeviceDetails struct {
	General struct {
		DeviceID     string `json:"device_id"`
		DeviceName   string `json:"device_name"`
		Model        string `json:"model"`
		Platform     string `json:"platform"`
		OSVersion    string `json:"os_version"`
		SerialNumber string `json:"serial_number"`
		AssetTag     string `json:"asset_tag"`
		BlueprintID  string `json:"blueprint_uuid"`
	} `json:"general"`
	MDM struct {
		Enabled      string   `json:"mdm_enabled"`
		Supervised   string   `json:"supervised"`
		LastCheckIn  string   `json:"last_check_in"`
		EnabledUsers []string `json:"mdm_enabled_user"`
	} `json:"mdm"`
}

// LibraryItemActivity represents activity for a library item.
type LibraryItemActivity struct {
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	Status       string `json:"status"`
	ActivityTime string `json:"activity_time"`
}

// LibraryItemStatus represents status for a library item on a device.
type LibraryItemStatus struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Status     string `json:"status"`
}

// Tag represents an Iru Tag.
type Tag struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

// CustomScript represents an Iru Custom Script library item.
type CustomScript struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	Active             bool   `json:"active"`
	ExecutionFrequency string `json:"execution_frequency"`
	Restart            bool   `json:"restart"`
	Script             string `json:"script"`
	RemediationScript  string `json:"remediation_script,omitempty"`
	ShowInSelfService  bool   `json:"show_in_self_service"`
}

// CustomProfile represents an Iru Custom Profile library item.
type CustomProfile struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Active        bool   `json:"active"`
	Profile       string `json:"profile,omitempty"`
	MDMIdentifier string `json:"mdm_identifier,omitempty"`
	RunsOnMac     bool   `json:"runs_on_mac"`
	RunsOnIPhone  bool   `json:"runs_on_iphone"`
	RunsOnIPad    bool   `json:"runs_on_ipad"`
	RunsOnTV      bool   `json:"runs_on_tv"`
	RunsOnVision  bool   `json:"runs_on_vision"`
}

// User represents an Iru User.
type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	IsArchived bool   `json:"is_archived"`
}

// PrismEntry represents a generic entry from a Prism endpoint.
type PrismEntry map[string]interface{}

// Vulnerability represents a vulnerability from Vulnerability Management.
type Vulnerability struct {
	CVEID              string   `json:"cve_id"`
	Severity           string   `json:"severity"`
	CVSSScore          float64  `json:"cvss_score"`
	FirstDetectionDate string   `json:"first_detection_date"`
	DeviceCount        int      `json:"device_count"`
	Status             string   `json:"status"`
	Software           []string `json:"software"`
}

// CustomApp represents an Iru Custom App library item.
type CustomApp struct {
	ID                     string `json:"id,omitempty"`
	Name                   string `json:"name"`
	FileKey                string `json:"file_key"`
	InstallType            string `json:"install_type"`
	InstallEnforcement     string `json:"install_enforcement"`
	UnzipLocation          string `json:"unzip_location,omitempty"`
	AuditScript            string `json:"audit_script,omitempty"`
	PreinstallScript       string `json:"preinstall_script,omitempty"`
	PostinstallScript      string `json:"postinstall_script,omitempty"`
	ShowInSelfService      bool   `json:"show_in_self_service"`
	SelfServiceCategoryID  string `json:"self_service_category_id,omitempty"`
	SelfServiceRecommended bool   `json:"self_service_recommended"`
	Active                 bool   `json:"active"`
	Restart                bool   `json:"restart"`
}

// InHouseApp represents an Iru In-House App library item (.ipa).
type InHouseApp struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	FileKey      string `json:"file_key"`
	RunsOnIPhone bool   `json:"runs_on_iphone"`
	RunsOnIPad   bool   `json:"runs_on_ipad"`
	RunsOnTV     bool   `json:"runs_on_tv"`
	Active       bool   `json:"active"`
}

// AuditEvent represents an audit log event.
type AuditEvent struct {
	ID              string      `json:"id"`
	Action          string      `json:"action"`
	OccurredAt      string      `json:"occurred_at"`
	ActorID         string      `json:"actor_id"`
	ActorType       string      `json:"actor_type"`
	TargetID        string      `json:"target_id"`
	TargetType      string      `json:"target_type"`
	TargetComponent string      `json:"target_component"`
	NewState        interface{} `json:"new_state"`
}

// Licensing represents tenant licensing information.
type Licensing struct {
	Counts struct {
		ComputersCount int `json:"computers_count"`
		IOSCount       int `json:"ios_count"`
		IPadOSCount    int `json:"ipados_count"`
		MacOSCount     int `json:"macos_count"`
		TVOSCount      int `json:"tvos_count"`
	} `json:"counts"`
	Limits struct {
		PlanType    string `json:"plan_type"`
		MaxDevices  int    `json:"max_devices"`
	} `json:"limits"`
	TenantOverLicenseLimit bool `json:"tenantOverLicenseLimit"`
}

// Threat represents a detected malware/pup threat.
type Threat struct {
	ThreatName         string `json:"threat_name"`
	Classification     string `json:"classification"`
	Status             string `json:"status"`
	DeviceName         string `json:"device_name"`
	DeviceID           string `json:"device_id"`
	DetectionDate      string `json:"detection_date"`
	FilePath           string `json:"file_path"`
	FileHash           string `json:"file_hash"`
	DeviceSerialNumber string `json:"device_serial_number"`
}

// BehavioralDetection represents a behavioral detection event.
type BehavioralDetection struct {
	ID             string `json:"id"`
	ThreatID       string `json:"threat_id"`
	Description    string `json:"description"`
	Classification string `json:"classification"`
	DetectionDate  string `json:"detection_date"`
	ThreatStatus   string `json:"threat_status"`
	DeviceInfo     struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		SerialNumber string `json:"serial_number"`
	} `json:"device_info"`
}

// DeviceSecretsALBC represents Activation Lock Bypass Code response.
type DeviceSecretsALBC struct {
	UserBasedALBC   string `json:"user_based_albc"`
	DeviceBasedALBC string `json:"device_based_albc"`
}

// DeviceSecretsFileVault represents FileVault Recovery Key response.
type DeviceSecretsFileVault struct {
	Key string `json:"key"`
}

// DeviceSecretsUnlockPin represents Unlock Pin response.
type DeviceSecretsUnlockPin struct {
	Pin string `json:"pin"`
}

// DeviceSecretsRecoveryLock represents Recovery Lock Password response.
type DeviceSecretsRecoveryLock struct {
	RecoveryPassword string `json:"recovery_password"`
}

// BlueprintLibraryItem represents a library item assigned to a blueprint.
type BlueprintLibraryItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ADEDevice represents an Iru ADE Device.
type ADEDevice struct {
	ID            string `json:"device_id,omitempty"`
	SerialNumber  string `json:"serial_number"`
	Model         string `json:"model"`
	Description   string `json:"description"`
	AssetTag      string `json:"asset_tag"`
	Color         string `json:"color"`
	BlueprintID   string `json:"blueprint_id"`
	UserID        string `json:"user_id"`
	DEPAccount    string `json:"dep_account"`
	DeviceFamily  string `json:"device_family"`
	OS            string `json:"os"`
	ProfileStatus string `json:"profile_status"`
	IsEnrolled    bool   `json:"is_enrolled"`
	UseBlueprintRouting bool `json:"use_blueprint_routing"`
}

// BlueprintRouting represents the Blueprint Routing settings.
type BlueprintRouting struct {
	EnrollmentCode struct {
		Code     string `json:"code"`
		IsActive bool   `json:"is_active"`
	} `json:"enrollment_code"`
}

// BlueprintRoutingActivity represents an activity event for Blueprint Routing.
type BlueprintRoutingActivity struct {
	ID           int                    `json:"id"`
	ActivityTime string                 `json:"activity_time"`
	ActivityType string                 `json:"activity_type"`
	User         map[string]interface{} `json:"user,omitempty"`
	DeviceID     string                 `json:"device_id,omitempty"`
	Details      map[string]interface{} `json:"details"`
}

// BlueprintRoutingActivityList represents a list of activity events for Blueprint Routing.
type BlueprintRoutingActivityList struct {
	Count    int                        `json:"count"`
	Next     string                     `json:"next"`
	Previous string                     `json:"previous"`
	Results  []BlueprintRoutingActivity `json:"results"`
}

// DeviceActivity represents a device activity event.
type DeviceActivity struct {
	ID               int                    `json:"id"`
	CreatedAt        string                 `json:"created_at"`
	ActionType       string                 `json:"action_type"`
	Details          map[string]interface{} `json:"details"`
	Computer         map[string]interface{} `json:"computer"`
	Blueprint        map[string]interface{} `json:"blueprint,omitempty"`
	User             map[string]interface{} `json:"user,omitempty"`
	BlueprintRouting bool                   `json:"blueprint_routing"`
}

// DeviceActivityList represents a list of device activity events.
type DeviceActivityList struct {
	DeviceID string           `json:"device_id"`
	Activity struct {
		Count    int              `json:"count"`
		Next     string           `json:"next"`
		Previous string           `json:"previous"`
		Results  []DeviceActivity `json:"results"`
	} `json:"activity"`
}

// DeviceCommand represents an MDM command sent to a device.
type DeviceCommand struct {
	UUID          string                 `json:"uuid"`
	CommandType   string                 `json:"command_type"`
	Status        int                    `json:"status"`
	DateRequested string                 `json:"date_requested"`
	DateCompleted string                 `json:"date_completed"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DeviceCommandList represents a list of device commands.
type DeviceCommandList struct {
	DeviceID string `json:"device_id"`
	Commands struct {
		Count    int             `json:"count"`
		Next     string          `json:"next"`
		Previous string          `json:"previous"`
		Results  []DeviceCommand `json:"results"`
	} `json:"commands"`
}

// BlueprintTemplate represents a blueprint template.
type BlueprintTemplate struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// BlueprintTemplateCategory represents a category of blueprint templates.
type BlueprintTemplateCategory struct {
	ID        int                 `json:"id"`
	Name      string              `json:"name"`
	Templates []BlueprintTemplate `json:"templates"`
}

// BlueprintTemplateList represents the response from the templates endpoint.
type BlueprintTemplateList struct {
	Count   int                         `json:"count"`
	Results []BlueprintTemplateCategory `json:"results"`
}

// PrismExport represents a Prism export job.
type PrismExport struct {
	ID        string `json:"id"`
	Status    string `json:"status"` // success, processing, failed
	Category  string `json:"category"`
	SignedURL string `json:"signed_url"`
}

// PrismCount represents the count of items in a Prism category.
type PrismCount struct {
	Count int `json:"count"`
}
