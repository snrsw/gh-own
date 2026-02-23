# Add Config File for Query Management

## TDD Test Plan

### Phase 1: Default query constants

- [x] `TestDefaultPRQueries_ContainsExpectedKeys` — 4 keys: created, assigned, participatedUser, reviewRequested; each contains `{user}`
- [ ] `TestDefaultIssueQueries_ContainsExpectedKeys` — 3 keys: created, assigned, participatedUser; each contains `{user}`

### Phase 2: ResolveQueries

- [ ] `TestResolveQueries_ReplacesUserPlaceholder` — `author:{user}` + `"octocat"` → `author:octocat`
- [ ] `TestResolveQueries_MultipleOccurrences` — `involves:{user} -author:{user}` → both replaced
- [ ] `TestResolveQueries_NoPlaceholder` — query without `{user}` passes through unchanged

### Phase 3: MergeQueries

- [ ] `TestMergePRQueries_NilOverride_ReturnsDefaults` — nil → copy of DefaultPRQueries
- [ ] `TestMergePRQueries_PartialOverride` — override `created` only, other 3 use defaults
- [ ] `TestMergePRQueries_FullOverride` — all 4 keys overridden
- [ ] `TestMergeIssueQueries_NilOverride_ReturnsDefaults` — nil → copy of DefaultIssueQueries
- [ ] `TestMergeIssueQueries_PartialOverride` — override `assigned` only, other 2 use defaults

### Phase 4: Key normalization

- [ ] `TestNormalizeKeys` — `participated` → `participatedUser`, `review_requested` → `reviewRequested`, passthrough for already-correct keys

### Phase 5: LoadFromPath

- [ ] `TestLoadFromPath_FileNotFound_ReturnsEmptyConfig` — no error, empty Config
- [ ] `TestLoadFromPath_ValidYAML_ParsesPRQueries` — temp YAML → correct Config.PR.Queries with normalized keys
- [ ] `TestLoadFromPath_ValidYAML_ParsesIssueQueries` — temp YAML → correct Config.Issue.Queries
- [ ] `TestLoadFromPath_InvalidYAML_ReturnsError` — malformed YAML → error
- [ ] `TestLoadFromPath_PartialConfig_OnlyPR` — only `pr:` section → Issue.Queries is nil

### Phase 6: DefaultPath

- [ ] `TestDefaultPath_UsesXDGConfigHome` — set env → `$XDG_CONFIG_HOME/gh-own/config.yaml`
- [ ] `TestDefaultPath_FallsBackToHomeDotConfig` — unset env → `~/.config/gh-own/config.yaml`

### Phase 7: Tidy — thread config through existing code

- [ ] Update `SearchPRs` signature: `SearchPRs(client, entries map[string]string)` — remove internal query construction
- [ ] Update `SearchIssues` signature: `SearchIssues(client, entries map[string]string)` — same
- [ ] Update `TestSearchPRs_EmptyUsername` → pass empty entries map
- [ ] Update `TestSearchIssues_EmptyUsername` → pass empty entries map
- [ ] Update `cmd/pr.go` — load config, merge, resolve, pass to `SearchPRs`
- [ ] Update `cmd/issue.go` — load config, merge, resolve, pass to `SearchIssues`
