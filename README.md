# 🤖 aiContext

Lightweight, zero-dependency Go CLI to bootstrap high-signal AI instruction context files (`AGENTS.md` & `CLAUDE.md`) in any project.

---

### **🚀 Getting Started (No Go needed)**

1. Go to `github.com/yourusername/aiContext/releases`
2. Download the binary for your OS and architecture.
3. Run the following setup:

**macOS / Linux:**
```bash
# make executable and move to PATH
chmod +x aiContext
mv aiContext /usr/local/bin/

# run global setup (once)
aiContext setup

# run per-project setup
cd myproject
aiContext init
```

**Windows:**
Move the `aiContext.exe` binary to any folder in your system `PATH` (e.g. a custom directory added to your PATH environment variables, or `%USERPROFILE%\bin`), then run `aiContext setup` globally and `aiContext init` per-project.

---

### **🛠️ How It Works**

* **`aiContext setup`**: Installs default instruction templates (`AGENTS.md` and `CLAUDE.md`) to your home configuration directory (`~/.config/aiContext/templates/`). You can customize these templates globally.
* **`aiContext init`**: Copies the templates to your current directory, replacing `{{PROJECT_NAME}}` with the current directory's base name. Prevents accidental overwrites.
