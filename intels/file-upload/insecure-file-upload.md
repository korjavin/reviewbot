---
id: file-upload-insecure
title: Insecure File Upload
severity: critical
tags: [file-upload, rce, path-traversal, content-type, validation]
taxonomy: security/web/file-upload
references:
  - https://owasp.org/www-community/vulnerabilities/Unrestricted_File_Upload
  - https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html
---

# Insecure File Upload

## Description

Insecure file upload is a critical vulnerability that can lead to remote code execution, stored XSS, path traversal, and server-side request forgery. Attackers upload web shells disguised as images, malicious scripts, or files with traversal paths in filenames.

Common attacks:
- Upload PHP/ASPX/JSP web shell → access uploaded file URL → RCE
- Bypass extension check with double extension (`shell.php.jpg`), null byte (`shell.php%00.jpg`), case variation (`Shell.PHP`)
- SVG upload with embedded JavaScript → stored XSS when rendered
- Upload with traversal filename → overwrite critical files

## Vulnerable Pattern

```python
# BAD — trusting Content-Type header and extension from client
@app.post("/upload")
async def upload(file: UploadFile, user=Depends(get_current_user)):
    if file.content_type not in ["image/jpeg", "image/png"]:
        raise HTTPException(400, "Invalid type")
    # BAD: content-type is user-controlled, extension not validated
    filename = file.filename  # could be "shell.php" or "../../../etc/cron.d/backdoor"
    with open(f"/app/uploads/{filename}", "wb") as f:
        f.write(await file.read())
    return {"url": f"/uploads/{filename}"}
```

## Secure Pattern

```python
import magic  # python-magic for content inspection
import uuid, os
from pathlib import Path

UPLOAD_DIR = Path("/app/uploads").resolve()
ALLOWED_EXTENSIONS = {".jpg", ".jpeg", ".png", ".gif", ".pdf", ".docx"}
ALLOWED_MIME_TYPES = {"image/jpeg", "image/png", "image/gif", "application/pdf"}
MAX_FILE_SIZE = 5 * 1024 * 1024  # 5MB

@app.post("/upload")
async def upload(file: UploadFile, user=Depends(get_current_user)):
    # 1. Check file size
    content = await file.read()
    if len(content) > MAX_FILE_SIZE:
        raise HTTPException(400, "File too large")

    # 2. Inspect magic bytes (not Content-Type header)
    mime = magic.from_buffer(content, mime=True)
    if mime not in ALLOWED_MIME_TYPES:
        raise HTTPException(400, "Invalid file type")

    # 3. Validate extension (double-check)
    original_ext = Path(file.filename).suffix.lower()
    if original_ext not in ALLOWED_EXTENSIONS:
        raise HTTPException(400, "Invalid extension")

    # 4. Generate safe, random filename (no path traversal)
    safe_filename = f"{uuid.uuid4()}{original_ext}"
    dest = UPLOAD_DIR / safe_filename

    # 5. Write to non-executable directory
    dest.write_bytes(content)

    # 6. Return URL that does NOT execute files
    return {"url": f"/static/uploads/{safe_filename}"}
```

## Checks to Generate

- Grep for file uploads that use `file.filename` directly as save path — path traversal risk.
- Flag uploads where MIME type validation uses `Content-Type` header (user-controlled) instead of magic bytes.
- Grep for `open(f"{upload_dir}/{filename}")` without `.resolve()` and prefix check.
- Flag upload directories that are also web-accessible and executable (same as webroot).
- Grep for SVG upload handling without XSS sanitization — SVG can contain JavaScript.
- Flag missing file size limits on upload endpoints.
- Check for zip/tar extraction after upload without zip-slip protection.
