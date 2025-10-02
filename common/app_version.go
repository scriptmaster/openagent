package common

// AppVersion is the current application version - hardcoded for deployment tracking
const AppVersion = "1.3.1.134"

// GetAppVersion returns the current application version
func GetAppVersion() string {
	return AppVersion
}
