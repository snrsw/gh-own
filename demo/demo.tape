Output demo/demo.gif

Set FontSize 30
Set Width 2400
Set Height 1200
Set Theme "Github"
Set Shell "bash"
Env PS1 ">> "

Type "gh own --demo"
Sleep 300ms
Enter
Sleep 1.5s

# Navigate through all PR tabs
Tab
Sleep 600ms
Tab
Sleep 600ms
Tab
Sleep 600ms

# Open selected item — shows "→ Opening ... in browser…" status
Enter
Sleep 2.5s

# Navigate back to Created tab
Shift+Tab
Sleep 100ms
Shift+Tab
Sleep 100ms
Shift+Tab
Sleep 500ms

# Filter demo
Type "/"
Sleep 500ms
Type "ci"
Sleep 1s
Escape
Sleep 500ms

Ctrl+C
Sleep 500ms

# Issues demo
Type "gh own issue --demo"
Sleep 300ms
Enter
Sleep 1.5s

# Navigate through all issue tabs
Tab
Sleep 600ms
Tab
Sleep 600ms

# Open selected item — shows "→ Opening ... in browser…" status
Enter
Sleep 2.5s

# Navigate back to Created tab
Shift+Tab
Sleep 100ms
Shift+Tab
Sleep 500ms

Ctrl+C
