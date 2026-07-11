# dbbee 🛠️🐝

`dbbee` is a lightweight, **CGO-free**, fast, and interactive Terminal User Interface (TUI) browser for SQLite databases, written in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

It is designed for developers who prefer keyboard-centric terminal workflows (Vim, multiplexers, CLI tools) and want to inspect schemas, browse tables, and edit data instantly without leaving the command line.

---

## ✨ Features

- **Zero Dependencies & CGO-Free**: Compiled to a single standalone binary using the pure Go SQLite parser (`modernc.org/sqlite`). Cross-compiles easily to Windows, macOS, and Linux without a C compiler toolchain.
- **Interactive Sidebar Navigation**: List all user tables, views, and row counts dynamically. Supports real-time fuzzy filtering (press `/` to filter).
- **Paginated Data Grid**: Aligned grid view that auto-calculates column widths based on terminal dimensions. Supports horizontal scrolling for tables with many columns and vertical paging (PgUp/PgDn).
- **Interactive Data Editing**:
  - Press **`n`** on the Grid to insert a new row.
  - Press **`enter`** on a cell to edit its value directly.
  - Press **`d`** on a row to delete it.
- **Schema Inspector**: Switch to schema mode (press `s`) to view the table columns metadata (Type, Primary Key, Not Null, Defaults) and syntax-highlighted raw `CREATE TABLE` DDL.
- **Multi-line SQL Query Editor**: Write and run custom SQLite queries. Prepend `EXPLAIN QUERY PLAN` automatically by pressing **`Ctrl+E`**. Supports suggestion autocomplete via **`Ctrl+Space`** / **`Ctrl+L`**.
- **Clipboard Exporter**: Quickly export currently viewed data page to the system clipboard as a CSV format (press `c` in the grid view).
- **Safe Read-Only Fallback**: Automatically opens locked or in-use databases in read-only mode to prevent blocking writes of other running processes.

---

## 🎹 Navigation & Hotkeys Quick Reference

| Mode / Tab | Key | Action |
| :--- | :--- | :--- |
| **Global** | `Tab` | Switch focus forward (Sidebar ➡️ Grid ➡️ Editor ➡️ Sidebar) |
| | `Shift+Tab` | Switch focus backward (Sidebar ⬅️ Editor ⬅️ Grid ⬅️ Sidebar) |
| | `q` / `Ctrl+c` | Exit dbbee |
| **Sidebar** | `j` / `k` (or `↓`/`↑`) | Navigate tables list |
| | `/` | Activate real-time table filter |
| | `Esc` | Cancel / Reset active table filter |
| | `Enter` | Load selected table data and switch focus to Grid |
| | `s` | View structural schema DDL for selected table |
| **Data Grid** | `h` / `l` (or `←`/`→`) | Navigate columns horizontally |
| | `j` / `k` (or `↓`/`↑`) | Navigate rows vertically |
| | `PgUp` / `PgDn` | Navigate page backward / forward |
| | `enter` | Edit the currently selected cell value |
| | `n` | Create a new row in the active table |
| | `d` | Delete the currently selected row (with y/n confirmation) |
| | `c` | Copy viewed page data to system clipboard as CSV |
| | `s` | Toggle back to table data view from Schema View |
| **Query Editor** | `Ctrl+J` / `Ctrl+Enter`| Execute the current SQL query |
| | `Ctrl+E` | Run `EXPLAIN QUERY PLAN` on current query |
| | `Ctrl+Space` / `Ctrl+L`| Autocomplete SQL keywords or table name suggestions |
| | `Up` / `Down` | Browse historical executed queries (when cursor is on first/last line) |

---

## 🚀 Installation & Build

### Prerequisites
- Go 1.21 or higher installed on your system.

### Build from Source
Clone the repository and build the binary:

```bash
# Clone the repository
git clone https://github.com/SiamAlSobari/Tui.git
cd Tui

# Build the executable
go build -o bin/dbbee ./cmd/dbbee
```

### Install Globally (Go Developers)

Anyone who has the Go toolchain installed can install `dbbee` globally with a single command:

```bash
go install github.com/SiamAlSobari/dbbee/cmd/dbbee@latest
```

Alternatively, to install from the local repository directory:

```bash
go install ./cmd/dbbee
```

---

## 💡 Usage

Run the executable and provide the path to your SQLite database file:

```bash
# Open database
dbbee path/to/database.db

# Open in read-only mode explicitly
dbbee -ro path/to/database.db
```

### Try with Sample Database
A sample database is populated with tables (`users`, `posts`, `comments`) and a view (`user_post_counts`) under `data/sample.db`. You can run:

```bash
go run ./cmd/dbbee/ data/sample.db
```

---

## 🧪 Testing

The codebase is fully covered by automated unit tests. To run tests for database clients, exporters, and TUI components:

```bash
go test -v ./...
```
