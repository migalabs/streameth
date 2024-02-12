package utils

const (
	PrysmClient      = "Prysm"
	LighthouseClient = "Lighthouse"
	TekuClient       = "Teku"
	NimbusClient     = "Nimbus"
	LodestarClient   = "Lodestar"
	GrandineClient   = "Grandine"
)

func CheckValidClientName(name string) bool {
	switch name {
	case PrysmClient:
		return true
	case LighthouseClient:
		return true
	case TekuClient:
		return true
	case NimbusClient:
		return true
	case LodestarClient:
		return true
	case GrandineClient:
		return true
	default:
		return false
	}
}
