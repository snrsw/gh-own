# Plan: Add `c` key binding for PR checkout

## TDD steps

- [x] 1. `WithCheckout` option sets `checkoutEnabled` on Model
- [x] 2. `c` key returns checkout command when enabled and item selected
- [x] 3. `c` key is no-op when checkout is disabled
- [x] 4. `c` key is ignored during filtering
- [x] 5. `c` key is no-op when no item selected (empty list)
- [x] 6. `checkoutExecCommand` builds correct `*exec.Cmd`
- [x] 7. Handle `checkoutFinishedMsg` in `Update`
- [x] 8. `helpView` shows `c checkout` when checkout enabled
- [x] 9. `helpView` hides `c checkout` when disabled or filtering
- [x] 10. Update existing `helpView` tests
- [x] 11. Wire `WithCheckout(true)` in `cmd/pr.go`
