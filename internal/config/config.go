package config

var DefaultPRQueries = map[string]string{
	"created":          "is:pr is:open author:{user}",
	"assigned":         "is:pr is:open assignee:{user}",
	"participatedUser": "is:pr is:open involves:{user} -author:{user} -assignee:{user} -review-requested:{user}",
	"reviewRequested":  "is:pr is:open review-requested:{user}",
}

var DefaultIssueQueries = map[string]string{
	"created":          "is:issue is:open author:{user}",
	"assigned":         "is:issue is:open assignee:{user}",
	"participatedUser": "is:issue is:open involves:{user} -author:{user} -assignee:{user}",
}
