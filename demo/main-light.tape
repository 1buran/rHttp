Output demo/main-light.gif

Set Shell "bash"
Set Theme "Belafonte Day"
Set FontSize 16
Set Padding 2
Set Margin 2
Set Width 1280
Set Height 1024
Set Framerate 60
Set TypingSpeed 150ms
Set PlaybackSpeed 0.5
Set CursorBlink false

Require rhttp
Type "rhttp -c demo/config-light.json" Enter

Sleep 3
Ctrl + h
Sleep 3

Ctrl + f
Sleep 2

Enter
Backspace 9
Copy "api.coinlore.net"
Paste
Enter
Copy "/api/tickers/"
Paste
Enter 9
Space
Enter
Space
Sleep 3
Ctrl + g
PageDown 10
Sleep 3
PageUp  10
Sleep 3
Ctrl + q
