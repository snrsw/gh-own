package config

import "strings"

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

func ResolveQueries(queries map[string]string, username string) map[string]string {
	resolved := make(map[string]string, len(queries))
	for key, query := range queries {
		resolved[key] = strings.ReplaceAll(query, "{user}", username)
	}
	return resolved
}
