Output demo/edit-json-payload.gif

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
Enter
Sleep 3
Ctrl + j
Sleep 3
Copy "{" Paste Enter
Copy '  "name": "Gabriel",' Paste Enter
Copy '  "password": "Swordfish",' Paste Enter
Copy '  "job": "self-employed"' Paste Enter
Alt + Enter
Sleep 3
Copy "}" Paste Enter
Alt + Enter
Sleep 3
Ctrl + j
Enter
Backspace 9
Copy "reqres.in" Paste Enter
Copy "/api/users" Paste Enter
Enter 8 Space
Enter Space
Ctrl + g
PageDown 10
Sleep 3
PageUp  10
Sleep 3
Ctrl + q
