Output demo/attach-file.gif

Set Shell "bash"
Set FontSize 16
Set Padding 2
Set Margin 2
Set Width 1280
Set Height 1024
Set Framerate 60
Set TypingSpeed 200ms
Set PlaybackSpeed 0.3
Set CursorBlink false

Require rhttp
Hide
Type "rhttp -c demo/config.json" Enter
Ctrl + f
Show
Sleep 2
Copy "POST" Paste Enter
Ctrl + p
Sleep 2
Copy "demo/payload.json"
Paste
Sleep 1
Enter 2
Backspace 9
Copy "reqres.in" Paste Enter
Copy "/api/users" Paste Enter
Enter 8 Space
Enter Space
Ctrl + g
Sleep 3
PageDown
Sleep 3
PageUp
Ctrl + q
