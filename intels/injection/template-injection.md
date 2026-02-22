---
id: injection-template
title: Server-Side Template Injection (SSTI)
severity: critical
tags: [injection, ssti, template, rce, jinja2, twig]
taxonomy: security/injection/template
references:
  - https://owasp.org/www-community/attacks/Server-Side_Template_Injection
  - https://portswigger.net/web-security/server-side-template-injection
---

# Server-Side Template Injection (SSTI)

## Description

SSTI occurs when user input is embedded into a template that is then rendered server-side. Template engines (Jinja2, Twig, Freemarker, Velocity, Pebble) execute embedded expressions. Attackers can use template syntax to execute arbitrary code, read files, or exfiltrate environment variables and secrets.

Detection payload: `{{7*7}}` → if response shows `49`, SSTI is confirmed.

## Vulnerable Pattern

```python
# BAD — Jinja2: user input rendered as template (Python/Flask)
from jinja2 import Template

@app.route("/greet")
def greet():
    name = request.args.get("name", "User")
    template = f"Hello, {name}!"
    return Template(template).render()
    # Payload: name={{config}} → leaks Flask config
    # Payload: name={{''.__class__.__mro__[1].__subclasses__()}} → RCE path

# Also vulnerable: Flask render_template_string with user data
return render_template_string(f"<h1>Hello {name}!</h1>")
```

```java
// BAD — Freemarker with user-supplied template
Configuration cfg = new Configuration();
Template template = new Template("name", new StringReader(userInput), cfg);
template.process(dataModel, out);
// Payload: <#assign ex="freemarker.template.utility.Execute"?new()>${ex("id")}
```

## Secure Pattern

```python
# GOOD — Jinja2: separate template from user data (use variables, not f-strings)
from jinja2 import Environment, select_autoescape

env = Environment(autoescape=select_autoescape())
template = env.from_string("Hello, {{ name }}!")  # template is a literal
return template.render(name=user_name)  # user data is a variable, not template code

# GOOD — Flask: use file templates, pass user data as context
return render_template("greet.html", name=user_name)
```

## Checks to Generate

- Grep for `Template(f"`, `render_template_string(f"`, `Template(user_` — user input in template string.
- Flag Jinja2 `Environment` with `autoescape=False` or no autoescape setting.
- Grep for Freemarker/Velocity/Twig template engines accepting user-supplied template strings.
- Grep for `${user_input}` in server-side template files where `user_input` comes from request.
- Flag pystache/handlebars/mustache template rendering with user-supplied template content.
- Check for SSTI in error message templates that include user-supplied identifiers.
