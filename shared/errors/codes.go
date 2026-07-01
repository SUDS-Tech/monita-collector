package apperr

const (
	
	CodeAgentUnknown        = "AGENT_UNKNOWN"
	CodeAgentRevoked        = "AGENT_REVOKED"
	CodeAgentExpired        = "AGENT_EXPIRED"
	CodeAgentFrozen         = "AGENT_FROZEN"
	CodeClockSkew           = "CLOCK_SKEW"
	CodeReplayDetected      = "REPLAY_DETECTED"
	CodeSignatureMismatch   = "SIGNATURE_MISMATCH"
	CodeFingerprintMismatch = "FINGERPRINT_MISMATCH"

	
	CodeRateLimited    = "RATE_LIMITED"
	CodePayloadTooLarge = "PAYLOAD_TOO_LARGE"

	
	CodeEmailTaken         = "EMAIL_TAKEN"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"

	
	CodeSubscriptionRequired = "SUBSCRIPTION_REQUIRED"
	CodeOrgSuspended         = "ORG_SUSPENDED"
)