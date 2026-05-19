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

Push a version tag to trigger the automatic release workflow:

```bash
# every release
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions automatically builds and uploads binaries to GitHub Releases using GoReleaser.
