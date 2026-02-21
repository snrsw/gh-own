package gh

import "time"

type LatestActivity struct {
	Kind  string
	Login string
	At    string
}

func NewLatestActivity(commentLogin, commentAt, reviewLogin, reviewAt, reviewState, pushLogin, pushAt string) LatestActivity {
	var candidates []LatestActivity
	if commentLogin != "" {
		candidates = append(candidates, LatestActivity{Kind: "commented", Login: commentLogin, At: commentAt})
	}
	if reviewLogin != "" {
		candidates = append(candidates, LatestActivity{Kind: reviewKind(reviewState), Login: reviewLogin, At: reviewAt})
	}
	if pushLogin != "" {
		candidates = append(candidates, LatestActivity{Kind: "pushed", Login: pushLogin, At: pushAt})
	}
	return mostRecent(candidates)
}

func mostRecent(candidates []LatestActivity) LatestActivity {
	if len(candidates) == 0 {
		return LatestActivity{}
	}
	best := candidates[0]
	bestTime, _ := time.Parse(time.RFC3339, best.At)
	for _, c := range candidates[1:] {
		t, err := time.Parse(time.RFC3339, c.At)
		if err == nil && t.After(bestTime) {
			best = c
			bestTime = t
		}
	}
	return best
}

func reviewKind(state string) string {
	switch state {
	case "APPROVED":
		return "approved"
	case "CHANGES_REQUESTED":
		return "changes requested"
	default:
		return "dismissed"
	}
}
