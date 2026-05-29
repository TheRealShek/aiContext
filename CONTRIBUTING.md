# 💻 Developing aiContext

Guidelines and commands for maintainers of `aiContext`.

---

### **🔨 Development Setup**

Add the git remote origin and update the module path:

```bash
# once
git remote add origin https://github.com/yourusername/aiContext
# update go.mod module path to match ^ real username
```

---

### **🚀 Making a Release**

Releases are fully automated!

- **Automatic Releases**: Every push or merge to the `main` branch automatically increments the patch version (e.g., `v1.0.3` → `v1.0.4`), tags the commit, and publishes the new release via GitHub Actions and GoReleaser.
- **Manual Releases**: If you want to release a specific version (like a major or minor version bump, e.g., `v2.0.0`), tag the commit manually and push:

  ```bash
  git tag v2.0.0
  git push origin v2.0.0
  ```

  GitHub Actions will detect the manual tag, skip the automatic patch increment, and release the exact version specified.
