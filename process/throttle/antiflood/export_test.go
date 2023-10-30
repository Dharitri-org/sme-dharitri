package antiflood

import "github.com/Dharitri-org/sme-dharitri/process"

func (af *p2pAntiflood) Debugger() process.AntifloodDebugger {
	return af.debugger
}
