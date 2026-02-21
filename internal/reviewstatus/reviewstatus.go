package reviewstatus

import "github.com/charmbracelet/lipgloss"

type ReviewStatus int

const (
	ReviewStatusNone ReviewStatus = iota
	ReviewStatusApproved
	ReviewStatusChangesRequested
	ReviewStatusReviewRequired
)

func (s ReviewStatus) String() string {
	switch s {
	case ReviewStatusApproved:
		return "approved"
	case ReviewStatusChangesRequested:
		return "changes_requested"
	case ReviewStatusReviewRequired:
		return "review_required"
	default:
		return "none"
	}
}

func ParseReviewDecision(decision string) ReviewStatus {
	switch decision {
	case "APPROVED":
		return ReviewStatusApproved
	case "CHANGES_REQUESTED":
		return ReviewStatusChangesRequested
	case "REVIEW_REQUIRED":
		return ReviewStatusReviewRequired
	default:
		return ReviewStatusNone
	}
}

var (
	approvedStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#1A7F37"))
	changesRequestedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CF222E"))
	reviewRequiredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9A6700"))
)

func RenderReviewStatus(status ReviewStatus) string {
	switch status {
	case ReviewStatusApproved:
		return approvedStyle.Render("✔")
	case ReviewStatusChangesRequested:
		return changesRequestedStyle.Render("⊘")
	case ReviewStatusReviewRequired:
		return reviewRequiredStyle.Render("◇")
	default:
		return ""
	}
}
