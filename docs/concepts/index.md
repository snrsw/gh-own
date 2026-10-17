---
okf_version: "0.2"
---

# gh-own concepts

Concepts an agent needs before changing gh-own. Each file owns one concept and links to the others instead of repeating them. User-facing behavior is documented in the README.

## Tabs

* [PR tabs](/tabs/pr-tabs.md) - The six default tabs of `gh-own pr`, the query behind each, and how query keys route results into tab buckets.
* [Issue tabs](/tabs/issue-tabs.md) - The three default tabs of `gh-own issue`; issues have no conversation reading and no tab promotion.

## Pipelines

* [Search pipeline](/pipelines/search.md) - How a command turns config and flags into GitHub search queries, runs user and team searches concurrently, and merges the results.

## Configuration

* [Config file](/config/config-file.md) - The YAML file at `~/.config/gh-own/config.yaml` and the CLI flags that override or extend it.

## Data

* [Conversation](/data/conversation.md) - The tail of a PR's discussion (description, comments, reviews, review threads, commits) fetched with each search result.

## Rules

* [Attention](/rules/attention.md) - Decides whether a PR's conversation is waiting on the user, and why (mentioned you / replied to you / commented).
* [Truncation horizon](/rules/truncation-horizon.md) - The cutoff time before which conversation events are ignored, because the user's answer may lie outside the fetched tail.
* [Needs action promotion](/rules/needs-action-promotion.md) - Moves PRs whose conversation is waiting on the user into the Needs action tab and out of the other default tabs.

## UI

* [Activity line and sort order](/ui/activity-and-sort.md) - What the second line of each list item reports, and why the list is ordered by exactly that timestamp.
* [Status indicators](/ui/status-indicators.md) - The CI and review status symbols appended to each PR title, and the GitHub GraphQL values they map from.
