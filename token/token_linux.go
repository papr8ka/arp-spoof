package token

import (
	"os"
)

func IsAdmin() bool {
	return os.Geteuid() == 0
}
