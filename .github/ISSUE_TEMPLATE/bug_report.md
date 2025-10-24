---
name: Bug report
about: Create a report to help us improve
title: ''
labels: 'bug'
assignees: ''

---

## Describe the bug 🐛
A clear and concise description of what the bug is. The more information you provide, the better we can help.

## To Reproduce 📝
Steps to reproduce the behavior:

1. What command did you run?
2. What arguments did you use?
3. What flags did you use?

**Examples:**

```bash
skipctl manifests diff -p my/path/to/data --diff-format json
skipctl manifests render -p my/path/to/data
cat my/file | skipctl manifests validate -
```

## Expected behavior ✅
A clear and concise description of what you expected to happen.

## Actual behavior ❌
What actually happened instead?

## Environment 🖥️
- `skipctl` version: (run `skipctl --version`)
- OS: (e.g., macOS 14.0, Ubuntu 22.04)
- Shell: (e.g., bash, zsh)

## Logs 🧾
Paste relevant logs or errors here (remember to redact any secrets):

```
Your logs here
```

## Screenshots 📸
If applicable, add screenshots to help explain your problem:
- Terminal output and error messages
- GitHub workflow errors (if using `skipctl` in CI/CD)

## Additional context 🌐
Add any other context about the problem here.
