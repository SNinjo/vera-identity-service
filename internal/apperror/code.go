package apperror

// the error code format is <status>_<service_id>_<serial_number>

const (
	CodeUserNotFound        = "404_01_001"
	CodeInvalidOAuthCode    = "500_01_002"
	CodeInvalidOAuthIdToken = "500_01_003"
	CodeMissingUserInfo     = "500_01_004"
	CodeInvalidRefreshToken = "401_01_005"
	CodeInvalidAccessToken  = "401_01_006"
	CodeInvalidTokenIssuer  = "401_01_007"
	CodeInvalidAuthHeader   = "401_01_008"
	CodeUserAlreadyExists   = "409_01_009"
	CodeEmailAlreadyInUse   = "409_01_010"
	CodeUserNotAuthorized   = "403_01_011"
)
