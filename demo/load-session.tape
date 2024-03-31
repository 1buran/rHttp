Output demo/load-session.gif

Set Shell "bash"
Set FontSize 16
Set Padding 2
Set Margin 2
Set Width 1280
Set Height 1024
Set Framerate 60
Set TypingSpeed 150ms
Set PlaybackSpeed 0.4
Set CursorBlink false

Require go
Type "go run ." Enter

Sleep 3
Ctrl + f
Enter
Ctrl + l
Copy "demo/session.json"
Sleep 3
Paste
Sleep 3
Enter
Sleep 3
PageDown 5
Sleep 3
PageUp  5
Sleep 3
Ctrl + q
