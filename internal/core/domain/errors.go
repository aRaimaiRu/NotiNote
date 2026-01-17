package domain

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
)

// OAuth errors
var (
	ErrOAuthStateMismatch = errors.New("oauth state mismatch - possible CSRF attack")
	ErrOAuthCodeExchange  = errors.New("failed to exchange oauth code for token")
	ErrOAuthUserInfo      = errors.New("failed to get user info from oauth provider")
	ErrOAuthProviderError = errors.New("oauth provider returned an error")
)

// Note errors
var (
	ErrNoteNotFound      = errors.New("note not found")
	ErrInvalidNoteData   = errors.New("invalid note data")
	ErrUnauthorizedAccess = errors.New("unauthorized access to resource")
)

// Notification errors
var (
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrInvalidScheduleTime   = errors.New("schedule time must be in the future")
	ErrNotificationCancelled = errors.New("notification has been cancelled")
	ErrNotificationFailed    = errors.New("failed to send notification")
)

// Device errors
var (
	ErrDeviceNotFound      = errors.New("device not found")
	ErrInvalidDeviceToken  = errors.New("invalid device token")
	ErrNoActiveDevices     = errors.New("no active devices found for user")
	ErrFCMSendFailed       = errors.New("failed to send FCM notification")
)

// Reminder errors
var (
	ErrReminderAccessDenied = errors.New("access denied to this reminder")
)

// Generic errors
var (
	ErrInternalServer = errors.New("internal server error")
	ErrValidation     = errors.New("validation error")
	ErrNotImplemented = errors.New("feature not implemented")
)
