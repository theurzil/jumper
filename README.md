# jumper

Autojump-style directory jumper. Remembers directories you visit and lets you jump to them by typing a fragment of their name.

```
j proj      # cd into the most "frecent" directory matching "proj"
```

## How it works

`jumper` is a small Go binary that stores a frecency-ranked history of visited directories in `~/.local/share/jumper/history.csv`. Frecency combines visit frequency and recency, so directories you use often and recently rank higher than ones you visited once months ago.

A `cd` builtin can only change the working directory of the shell that runs it, no external process (including this one) can change its parent shell's directory. So `jumper` itself only tracks history and answers queries; a tiny shell function (`j`) does the actual `cd`, and a shell hook calls `jumper add` on every directory change.

## Install

```bash
git clone https://github.com/theurzil/jumper.git
cd jumper
make install
```

This builds the binary to `/usr/local/bin/jumper`, copies `jumper.sh` to `~/.local/share/jumper/jumper.sh`, and adds a `source` line to `~/.bashrc`. If `/usr/local/bin` isn't writable, you'll be prompted for your `sudo` password. Zsh users, add the same line to `~/.zshrc` manually:

```bash
source ~/.local/share/jumper/jumper.sh
```

Restart your shell (or `source ~/.bashrc` / `source ~/.zshrc`).

To uninstall: `make uninstall`

## Usage

| Command          | Description                                    |
|------------------|-------------------------------------------------|
| `j <term>`       | cd to the best-ranked directory matching `term` |
| `j`              | cd to `$HOME`                                   |
| `j --help`       | show `j` help                                   |
| `jumper add <path>` | manually record a visit to `path`            |
| `jumper query <term>` | print the best-matching path (no cd)       |
| `jumper list`    | print all tracked paths, ranked by frecency     |
| `jumper --help`  | show `jumper` help                              |

History is built automatically: every `cd` (via the shell hook) records the new directory.

## Frecency scoring

```
age < 1 hour   -> frequency * 4
age < 1 day    -> frequency * 2
age < 1 week   -> frequency * 0.5
older          -> frequency * 0.25
```

## Uninstall

```bash
make uninstall   # or: sudo rm /usr/local/bin/jumper
rm ~/.local/share/jumper/history.csv
# and remove the `source jumper.sh` line from your shell rc file
```
