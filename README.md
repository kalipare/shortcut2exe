# Shortcut File To Exe Converter

This lib is used to convert shortcut file to executable file easily.

Supported file types:

- `.lnk`, `.url`, or `.cda` files; will become `.exe` file
- `.desktop` will become unix/linux/macos executable file

### How to run

Basic usage:

```bash
# download this package
git clone https://github.com/kalipare/shortcut2exe
cd shortcut2exe

# or use the go run command
go run main.go "path/to/icon/dot/url/file"
go run main.go "D:\some\path\of\Assassin's Creed Origins.url"
  => result: executable "Assassin's Creed Origins.exe" is successfully generated
```
