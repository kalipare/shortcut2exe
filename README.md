# Shortcut File To Exe Converter

This lib is used to convert shortcut file to executable file easily.

Supported file types:

- `.lnk`, `.url`, or `.cda` files; will become `.exe` file
- `.desktop` will become unix/linux/macos executable file

### How to run

Basic usage:

```bash
# download this package
go get github.com/novalagung/shortcut2exe

# run the installed binary
shortcut2exe "path/to/icon/dot/url/file"

# or use the go run command
cd shortcut2exe
go run main.go "path/to/icon/dot/url/file"
```

Example:

```bash
go run main.go "D:\some\path\of\Assassin's Creed Origins.url"
  => result: executable "Assassin's Creed Origins.exe" is successfully generated
```
