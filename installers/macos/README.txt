Control de Investimentos for macOS
==================================

1. Drag "Control de Investimentos.app" to the Applications folder.
2. Double-click it to open it. The app is not signed by Apple, so the first
   time macOS will not let it open: go to System Settings > Privacy & Security
   and click "Open Anyway" (on older versions: right-click the app > Open).
   You can also run this in Terminal:
       xattr -dr com.apple.quarantine "/Applications/Control de Investimentos.app"
3. A Terminal window opens with the server, and the browser opens the app.
   To close it, press Ctrl+C in the Terminal.

Your data is stored in ~/Library/Application Support/Control de Investimentos
(investimentos.db). If you already had a database, copy it there before
opening the app.

The app is available in Galician and English: change the language with the
selector at the top of the page or in Settings.
