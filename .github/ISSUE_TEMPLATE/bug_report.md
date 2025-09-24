---
name: Bug report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Use DDEX file '...'
2. Call function '....'
3. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Code sample**
```go
// Minimal code sample that reproduces the issue
package main

import (
    "github.com/OpenAudio/ddex-proto"
)

func main() {
    // Your code here
}
```

**Error output**
```
Paste the full error output here
```

**DDEX file (if applicable)**
If the bug is related to parsing a specific DDEX file:
- [ ] I can share the DDEX file publicly
- [ ] I cannot share the DDEX file (please provide sanitized version)
- [ ] The issue occurs with official DDEX sample files

**Environment:**
 - OS: [e.g. macOS, Linux, Windows]
 - Go version: [e.g. 1.25.0]
 - ddex-proto version: [e.g. v1.0.0 or commit hash]
 - DDEX specification: [e.g. ERN v4.3.2, MEAD v1.1, PIE v1.0]

**Additional context**
Add any other context about the problem here, such as:
- Is this a regression from a previous version?
- Does this affect XML, JSON, or protobuf serialization?
- Are you using custom generation or standard library?