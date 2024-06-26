package token

import (
	"github.com/papr8ka/arp-spoof/logger"
	"go.uber.org/zap"
	"golang.org/x/sys/windows"
)

// IsAdmin checks if user has admin token
// From https://github.com/golang/go/issues/28804
func IsAdmin() bool {
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		logger.Logger.Error("SID error",
			zap.Error(err))
		return false
	}
	defer func() { _ = windows.FreeSid(sid) }()

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		logger.Logger.Error("token Membership Error",
			zap.Error(err))
		return false
	}

	return member
}
