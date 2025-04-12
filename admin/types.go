package admin

// MaintenanceLoginData holds data for the maintenance login template.
type MaintenanceLoginData struct {
	Error      string
	AdminEmail string
	AppVersion string
}

// MaintenanceConfigData holds data for the maintenance configuration template.
type MaintenanceConfigData struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string // Note: Consider security implications of displaying password
	DBName         string
	Error          string
	Success        string
	AdminEmail     string
	VersionMajor   int
	VersionMinor   int
	VersionPatch   int
	VersionBuild   int
	MigrationStart string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPassword   string // Note: Consider security implications of displaying password
	SMTPFrom       string
}
