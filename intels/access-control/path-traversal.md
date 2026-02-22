---
id: access-control-path-traversal
title: Path Traversal (Directory Traversal)
severity: high
tags: [access-control, path-traversal, file-read, lfi]
taxonomy: security/access-control/path-traversal
references:
  - https://owasp.org/www-community/attacks/Path_Traversal
  - https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html
---

# Path Traversal (Directory Traversal)

## Description

Path traversal allows attackers to read files outside the intended directory by using sequences like `../` in file parameters. Can expose `/etc/passwd`, application source code, configuration files with credentials, and private keys. Severe when combined with file upload (write traversal).

## Vulnerable Pattern

```python
# BAD — serving user-supplied filename from a directory
from pathlib import Path

UPLOAD_DIR = Path("/app/uploads")

@app.get("/files/{filename}")
def serve_file(filename: str):
    file_path = UPLOAD_DIR / filename
    return FileResponse(file_path)
    # Payload: filename = "../../etc/passwd" → reads /etc/passwd

# BAD — extracting ZIP/TAR without path check (zip slip)
import zipfile
def extract_archive(zip_path: str, dest_dir: str):
    with zipfile.ZipFile(zip_path) as zf:
        zf.extractall(dest_dir)  # zip entry: "../../../etc/cron.d/backdoor"
```

## Secure Pattern

```python
from pathlib import Path
import os

UPLOAD_DIR = Path("/app/uploads").resolve()

@app.get("/files/{filename}")
def serve_file(filename: str):
    # Resolve and verify path stays within allowed directory
    requested = (UPLOAD_DIR / filename).resolve()
    if not str(requested).startswith(str(UPLOAD_DIR)):
        raise HTTPException(403, "Access denied")
    if not requested.exists():
        raise HTTPException(404)
    return FileResponse(requested)

# GOOD — safe zip extraction with path validation
def extract_archive(zip_path: str, dest_dir: str):
    dest = Path(dest_dir).resolve()
    with zipfile.ZipFile(zip_path) as zf:
        for member in zf.namelist():
            target = (dest / member).resolve()
            if not str(target).startswith(str(dest)):
                raise ValueError(f"Zip slip detected: {member}")
        zf.extractall(dest)
```

## Checks to Generate

- Grep for `open(request.`, `FileResponse(path)`, `send_file(filename)` where filename comes from request.
- Flag `Path(base_dir) / user_input` without `.resolve()` and prefix check.
- Grep for `zipfile.extractall(` without member path validation — zip slip vulnerability.
- Flag `tarfile.extract(` without `filter=` parameter (Python 3.12+ safety filter).
- Check for `os.path.join(base, user_input)` — join with absolute path in user_input discards base.
- Grep for `../` in URL routing patterns without normalization.
