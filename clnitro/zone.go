package clnitro

const (
	InvalidAccessZone AccessZone = ""
	AnyAccessZone     AccessZone = "any"
	LanAccessZone     AccessZone = "lan"
)

type AccessZone string

func parseAccessZone(s string) AccessZone {
	switch s {
	case string(AnyAccessZone):
		return AnyAccessZone
	case string(LanAccessZone):
		return LanAccessZone
	default:
		return InvalidAccessZone
	}
}
