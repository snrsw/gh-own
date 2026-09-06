package gh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/config"
)

// SearchPRs runs every entry as a search. With conversation, each result also
// carries the tail of its discussion (see PRSearchNode.Conversation), at the
// cost of a larger response.
func SearchPRs(client *api.GraphQLClient, entries map[string]string, conversation bool) (*PRSearchResult, error) {
	if len(entries) == 0 {
		return &PRSearchResult{Custom: make(map[string][]PRSearchNode)}, nil
	}

	raw, err := Search(client, prSearchQuery(conversation), entries, parsePRSearchJSON)
	if err != nil {
		return nil, err
	}

	return parsePRSearchResult(raw)
}

// SearchPRsTeams searches the pull requests of every team; see SearchPRs for
// conversation.
func SearchPRsTeams(client *api.GraphQLClient, username string, teams []string, filters config.Filters, conversation bool) (*PRSearchResult, error) {
	if username == "" {
		return &PRSearchResult{Custom: make(map[string][]PRSearchNode)}, nil
	}

	if len(teams) == 0 {
		return &PRSearchResult{Custom: make(map[string][]PRSearchNode)}, nil
	}

	raw, err := Search(client, prSearchQuery(conversation), prTeamEntries(teams, filters), parsePRSearchJSON)
	if err != nil {
		return nil, err
	}

	return parsePRSearchResult(raw)
}

// prTeamEntries builds one query per team. The keys share the "participated"
// prefix so they route into the Participated bucket (see parsePRSearchResult).
// These queries are not user-configurable, so they always take the sort
// qualifier.
func prTeamEntries(teams []string, filters config.Filters) map[string]string {
	entries := make(map[string]string, len(teams))
	for i, team := range teams {
		entries[fmt.Sprintf("participatedTeam%d", i)] = fmt.Sprintf("is:pr is:open team:%s", team)
	}
	return config.ApplyFilters(entries, filters)
}

type PRSearchResult struct {
	Drafts          []PRSearchNode
	NeedsAction     []PRSearchNode
	ReadyToMerge    []PRSearchNode
	Waiting         []PRSearchNode
	Participated    []PRSearchNode
	ReviewRequested []PRSearchNode
	Custom          map[string][]PRSearchNode
}

func MergeSearchPRsResults(a, b *PRSearchResult) *PRSearchResult {
	custom := make(map[string][]PRSearchNode)
	for k, v := range a.Custom {
		custom[k] = v
	}
	for k, v := range b.Custom {
		custom[k] = append(custom[k], v...)
	}
	for k, v := range custom {
		custom[k] = deduplicatePRNodes(v)
	}

	merged := &PRSearchResult{
		Drafts:          deduplicatePRNodes(append(a.Drafts, b.Drafts...)),
		NeedsAction:     deduplicatePRNodes(append(a.NeedsAction, b.NeedsAction...)),
		ReadyToMerge:    deduplicatePRNodes(append(a.ReadyToMerge, b.ReadyToMerge...)),
		Waiting:         deduplicatePRNodes(append(a.Waiting, b.Waiting...)),
		Participated:    deduplicatePRNodes(append(a.Participated, b.Participated...)),
		ReviewRequested: deduplicatePRNodes(append(a.ReviewRequested, b.ReviewRequested...)),
		Custom:          custom,
	}
	return merged
}

func parsePRSearchJSON(data json.RawMessage) ([]PRSearchNode, error) {
	var sr struct {
		Nodes []prSearchRawNode `json:"nodes"`
	}
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, err
	}
	return parsePRSearchNodes(sr.Nodes), nil
}

// prSearchQuery builds the search query. Besides the fields shown in the list
// it fetches, with conversation, the tail of each pull request's discussion so
// that Conversation.Attention can tell whether it is waiting on the user, and
// without it only the single latest comment, review and commit that the
// activity line needs. The conversation page sizes are a trade-off: enough
// history to find the user's last activity and what followed it, small enough
// to keep the response of a search with fifty results reasonable.
func prSearchQuery(conversation bool) string {
	fields := prActivityFields
	if conversation {
		fields = prConversationFields
	}
	return `query($q: String!) {
	result: search(query: $q, type: ISSUE, first: 50) {
		nodes {
			... on PullRequest {
				number
				title
				url
				isDraft
				updatedAt
				createdAt
				reviewDecision
				author { login }
				repository { nameWithOwner }
` + fields + `
			}
		}
	}
}`
}

const prActivityFields = `
				commits(last: 1) {
					nodes {
						commit {
							statusCheckRollup { state }
							committedDate
							author { user { login } }
						}
					}
				}
				comments(last: 1) {
					nodes { author { login } createdAt }
				}
				reviews(last: 1) {
					nodes { author { login } submittedAt state }
				}`

const prConversationFields = `
				body
				commits(last: 5) {
					nodes {
						commit {
							statusCheckRollup { state }
							committedDate
							author { user { login } }
						}
					}
				}
				comments(last: 10) {
					nodes { author { __typename login } body createdAt }
				}
				reviews(last: 10) {
					nodes { author { __typename login } body submittedAt state }
				}
				reviewThreads(last: 10) {
					nodes {
						isResolved
						comments(last: 10) {
							nodes { author { __typename login } body createdAt }
						}
					}
				}`

func parsePRSearchResult(parsed map[string][]PRSearchNode) (*PRSearchResult, error) {
	defaultKeys := config.DefaultPRKeys()
	var drafts, needsAction, readyToMerge, waiting, participated []PRSearchNode
	custom := make(map[string][]PRSearchNode)

	for _, key := range sortedKeys(parsed) {
		nodes := parsed[key]
		switch {
		case strings.HasPrefix(key, "drafts"):
			drafts = append(drafts, nodes...)
		case strings.HasPrefix(key, "needsAction"):
			needsAction = append(needsAction, nodes...)
		case strings.HasPrefix(key, "readyToMerge"):
			readyToMerge = append(readyToMerge, nodes...)
		case strings.HasPrefix(key, "waiting"):
			waiting = append(waiting, nodes...)
		case strings.HasPrefix(key, "participated"):
			participated = append(participated, nodes...)
		case !defaultKeys[key]:
			custom[key] = nodes
		}
	}

	return &PRSearchResult{
		Drafts:          deduplicatePRNodes(drafts),
		NeedsAction:     deduplicatePRNodes(needsAction),
		ReadyToMerge:    deduplicatePRNodes(readyToMerge),
		Waiting:         deduplicatePRNodes(waiting),
		Participated:    deduplicatePRNodes(participated),
		ReviewRequested: deduplicatePRNodes(parsed["reviewRequested"]),
		Custom:          custom,
	}, nil
}

type PRSearchNode struct {
	Number         int
	Title          string
	URL            string
	IsDraft        bool
	UpdatedAt      string
	CreatedAt      string
	StatusState    string
	ReviewDecision string
	LatestActivity LatestActivity
	// Conversation is the recent discussion, see Conversation.Attention.
	Conversation Conversation
	// Attention is set once the pull request is found to be waiting on the
	// user; its zero value means it is not.
	Attention Attention
	Author    struct {
		Login string
	}
	Repository struct {
		NameWithOwner string
	}
}

func (p *PRSearchNode) CIStatus() cistatus.CIStatus {
	return cistatus.ParseState(p.StatusState)
}

func (p *PRSearchNode) RepositoryURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s", p.Repository.NameWithOwner)
}

type prSearchRawNode struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	Body           string `json:"body"`
	IsDraft        bool   `json:"isDraft"`
	UpdatedAt      string `json:"updatedAt"`
	CreatedAt      string `json:"createdAt"`
	ReviewDecision string `json:"reviewDecision"`
	Author         struct {
		Login string `json:"login"`
	} `json:"author"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Commits struct {
		Nodes []rawCommitNode `json:"nodes"`
	} `json:"commits"`
	Comments struct {
		Nodes []rawComment `json:"nodes"`
	} `json:"comments"`
	Reviews struct {
		Nodes []rawReview `json:"nodes"`
	} `json:"reviews"`
	ReviewThreads struct {
		Nodes []rawReviewThread `json:"nodes"`
	} `json:"reviewThreads"`
}

// rawActor is a GraphQL Actor. The typename tells a GitHub App ("Bot") from a
// person ("User"); the login of an App carries no "[bot]" suffix in GraphQL.
type rawActor struct {
	TypeName string `json:"__typename"`
	Login    string `json:"login"`
}

func (a rawActor) comment(body, at string) Comment {
	return Comment{Login: a.Login, IsBot: a.TypeName == "Bot", Body: body, At: at}
}

type rawComment struct {
	Author    rawActor `json:"author"`
	Body      string   `json:"body"`
	CreatedAt string   `json:"createdAt"`
}

type rawReview struct {
	Author      rawActor `json:"author"`
	Body        string   `json:"body"`
	SubmittedAt string   `json:"submittedAt"`
	State       string   `json:"state"`
}

type rawReviewThread struct {
	IsResolved bool `json:"isResolved"`
	Comments   struct {
		Nodes []rawComment `json:"nodes"`
	} `json:"comments"`
}

type rawCommitNode struct {
	Commit struct {
		StatusCheckRollup *struct {
			State string `json:"state"`
		} `json:"statusCheckRollup"`
		CommittedDate string `json:"committedDate"`
		Author        struct {
			User *struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"author"`
	} `json:"commit"`
}

func (c rawCommitNode) commit() Commit {
	cm := Commit{At: c.Commit.CommittedDate}
	if c.Commit.Author.User != nil {
		cm.Login = c.Commit.Author.User.Login
	}
	return cm
}

func parsePRSearchNodes(rawNodes []prSearchRawNode) []PRSearchNode {
	nodes := make([]PRSearchNode, 0, len(rawNodes))
	for _, n := range rawNodes {
		if n.Number == 0 {
			continue
		}
		node := PRSearchNode{
			Number:         n.Number,
			Title:          n.Title,
			URL:            n.URL,
			IsDraft:        n.IsDraft,
			UpdatedAt:      n.UpdatedAt,
			CreatedAt:      n.CreatedAt,
			ReviewDecision: n.ReviewDecision,
			Conversation:   n.conversation(),
		}
		node.Author.Login = n.Author.Login
		node.Repository.NameWithOwner = n.Repository.NameWithOwner

		// The lists are chronological, so the latest entry is the last one.
		if last := len(n.Commits.Nodes) - 1; last >= 0 && n.Commits.Nodes[last].Commit.StatusCheckRollup != nil {
			node.StatusState = n.Commits.Nodes[last].Commit.StatusCheckRollup.State
		}
		node.LatestActivity = n.latestActivity(node.Conversation)

		nodes = append(nodes, node)
	}
	return nodes
}

func (n prSearchRawNode) conversation() Conversation {
	c := Conversation{
		Author:    n.Author.Login,
		CreatedAt: n.CreatedAt,
		Body:      n.Body,
	}
	for _, cm := range n.Comments.Nodes {
		c.Comments = append(c.Comments, cm.Author.comment(cm.Body, cm.CreatedAt))
	}
	for _, r := range n.Reviews.Nodes {
		if r.State == "PENDING" {
			continue // Not submitted yet: only its author can see it.
		}
		c.Reviews = append(c.Reviews, Review{Comment: r.Author.comment(r.Body, r.SubmittedAt), State: r.State})
	}
	for _, th := range n.ReviewThreads.Nodes {
		thread := ReviewThread{IsResolved: th.IsResolved}
		for _, cm := range th.Comments.Nodes {
			thread.Comments = append(thread.Comments, cm.Author.comment(cm.Body, cm.CreatedAt))
		}
		c.Threads = append(c.Threads, thread)
	}
	for _, cm := range n.Commits.Nodes {
		c.Commits = append(c.Commits, cm.commit())
	}
	return c
}

// latestActivity picks the most recent comment, review, and push out of the
// conversation and lets NewLatestActivity choose between them.
func (prSearchRawNode) latestActivity(c Conversation) LatestActivity {
	var commentLogin, commentAt string
	if n := len(c.Comments); n > 0 {
		commentLogin, commentAt = c.Comments[n-1].Login, c.Comments[n-1].At
	}
	var reviewLogin, reviewAt, reviewState string
	if n := len(c.Reviews); n > 0 {
		reviewLogin, reviewAt, reviewState = c.Reviews[n-1].Login, c.Reviews[n-1].At, c.Reviews[n-1].State
	}
	var pushLogin, pushAt string
	if n := len(c.Commits); n > 0 && c.Commits[n-1].Login != "" {
		pushLogin, pushAt = c.Commits[n-1].Login, c.Commits[n-1].At
	}
	return NewLatestActivity(commentLogin, commentAt, reviewLogin, reviewAt, reviewState, pushLogin, pushAt)
}

func deduplicatePRNodes(nodes []PRSearchNode) []PRSearchNode {
	seen := make(map[string]bool)
	result := make([]PRSearchNode, 0, len(nodes))
	for _, n := range nodes {
		if seen[n.URL] {
			continue
		}
		seen[n.URL] = true
		result = append(result, n)
	}
	return result
}
