Output demo/attach-file.gif

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

Require rhttp
Type "rhttp" Enter

Sleep 3
Ctrl + f
Copy "POST" Paste Enter
Ctrl + p
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
