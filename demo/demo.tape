Output demo/demo.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Github"
Set Shell "bash"
Env PS1 ">> "

Type "go run . --demo"
Enter
Sleep 1s

# Open selected item in browser
Enter
Sleep 1s

# Navigate tabs
Tab
Sleep 500ms
Tab
Sleep 500ms
Tab
Sleep 500ms

# Filter
Type "/"
Type "ci"
Escape
Sleep 500ms

# Back to first tab
Shift+Tab
Shift+Tab
Shift+Tab
Sleep 500ms

Ctrl+C
