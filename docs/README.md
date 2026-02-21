# ReviewBot Documentation & Blog

A GitHub Pages site documenting ReviewBot's architecture and design decisions.

## Local Development

### Prerequisites
- Ruby (macOS includes it)
- Bundler (`gem install bundler`)

### Setup

```bash
cd docs

# Install dependencies
bundle install

# Run Jekyll server
bundle exec jekyll serve
```

The site will be available at `http://localhost:4000`

### File Structure

```
docs/
├── _config.yml              # Jekyll configuration
├── _layouts/                # HTML templates
│   ├── default.html        # Base layout
│   └── post.html           # Blog post layout
├── _posts/                 # Blog posts (markdown)
├── assets/
│   ├── css/style.css       # Custom styles
│   └── images/             # Screenshots & diagrams
├── index.md                # Homepage
└── 404.md                  # 404 page
```

## Writing Blog Posts

Create a file in `_posts/` with the naming convention:
```
YYYY-MM-DD-title-slug.md
```

Example frontmatter:
```yaml
---
layout: post
title: "Your Post Title"
date: 2026-02-21 08:00:00 -0000
categories: architecture design
excerpt: "Short description of the post"
---
```

## Images

Place screenshots and diagrams in `assets/images/`:
- Use descriptive filenames
- PNG format preferred
- Optimize before uploading
- Images are clickable for zoom

## Deployment

GitHub Pages automatically builds and deploys on push to main:

1. Commit changes to `docs/`
2. Push to `main` branch
3. GitHub Actions builds the site
4. Available at `https://github.com/korjavin/reviewbot/tree/main/docs`

## Customization

### Colors
Edit `:root` variables in `assets/css/style.css`:
```css
:root {
  --primary: #2c3e50;
  --accent: #3498db;
  --light-bg: #f8f9fa;
  /* ... */
}
```

### Theme
The site uses a custom light theme with:
- Minimalist design
- Responsive layout (mobile-friendly)
- Image zoom functionality
- Syntax highlighting for code

## SEO

The site includes:
- Automatic sitemap generation (Jekyll plugin)
- RSS feed (`/feed.xml`)
- Meta tags for social sharing
- robots.txt for search engines

## Troubleshooting

### Jekyll not building?
```bash
bundle exec jekyll clean
bundle exec jekyll build
```

### Images not showing?
Check image paths use `/assets/images/filename.png`

### CSS not loading?
Clear browser cache or use hard refresh (Cmd+Shift+R)

---

For more info: [Jekyll Docs](https://jekyllrb.com/docs/)
