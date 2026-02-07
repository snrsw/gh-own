package cistatus

import "github.com/charmbracelet/lipgloss"

type CIStatus int

const (
	CIStatusNone CIStatus = iota
	CIStatusSuccess
	CIStatusFailure
	CIStatusPending
)

func (s CIStatus) String() string {
	switch s {
	case CIStatusSuccess:
		return "success"
	case CIStatusFailure:
		return "failure"
	case CIStatusPending:
		return "pending"
	default:
		return "none"
	}
}

func ParseState(state string) CIStatus {
	switch state {
	case "SUCCESS":
		return CIStatusSuccess
	case "FAILURE", "ERROR":
		return CIStatusFailure
	case "PENDING", "EXPECTED":
		return CIStatusPending
	default:
		return CIStatusNone
	}
}

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1A7F37"))
	failureStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CF222E"))
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9A6700"))
	noneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6E7781"))
)

func RenderCIStatus(status CIStatus) string {
	switch status {
	case CIStatusSuccess:
		return successStyle.Render("✓")
	case CIStatusFailure:
		return failureStyle.Render("✗")
	case CIStatusPending:
		return pendingStyle.Render("●")
	default:
		return noneStyle.Render("-")
	}
}
