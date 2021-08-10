package region

import (
	"github.com/azazeal/fly/pkg/env"
	"github.com/azazeal/flycast/internal/common"
)

// Name is shorthand for global ? "global" : env.Region().
func Name(global bool) string {
	if global {
		return ""
	}

	return env.Region()
}

// Alias is shorthand for global ? "global" : "local".
func Alias(global bool) string {
	if global {
		return "global"
	}

	return "local"
}

// PeerComponent is shorthand for global ? HCRefreshGlobalComponent :
// HCRefreshLocalComponent.
func PeerComponent(global bool) string {
	if global {
		return common.HCRefreshGlobalComponent
	}

	return common.HCRefreshLocalComponent
}

// WireComponent is shorthand for global ? HCWireGlobal : HCWireLocal.
func WireComponent(global bool) string {
	if global {
		return common.HCWireGlobal
	}

	return common.HCWireLocal
}
