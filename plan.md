# Plan: Add `c` key binding for PR checkout

## TDD steps

- [x] 1. `WithCheckout` option sets `checkoutEnabled` on Model
- [ ] 2. `c` key returns checkout command when enabled and item selected
- [ ] 3. `c` key is no-op when checkout is disabled
- [ ] 4. `c` key is ignored during filtering
- [ ] 5. `c` key is no-op when no item selected (empty list)
- [ ] 6. `checkoutExecCommand` builds correct `*exec.Cmd`
- [ ] 7. Handle `checkoutFinishedMsg` in `Update`
- [ ] 8. `helpView` shows `c checkout` when checkout enabled
- [ ] 9. `helpView` hides `c checkout` when disabled or filtering
- [ ] 10. Update existing `helpView` tests
- [ ] 11. Wire `WithCheckout(true)` in `cmd/pr.go`
