# Pensieve - Static Site Generator

A minimal static site generator for your thoughts, written in Go.

## Structure

```
pensieve/
├── main.go              # The generator
├── go.mod               # Go dependencies
├── style.css            # CSS stylesheet
├── templates/           # HTML templates
│   ├── index.html
│   ├── all.html
│   ├── thought.html
│   ├── tag.html
│   ├── archive-index.html
│   └── archive-month.html
├── thoughts/            # Your markdown thoughts
│   └── *.md
└── output/              # Generated HTML (created by build)
```

## Markdown Format

Create files in `thoughts/` with this format:

```markdown
---
tags: tag1, tag2, tag3
date: 2025-02-23
---

Your thought content here. Can be multiple paragraphs.

Markdown is supported: **bold**, *italic*, [links](https://example.com), etc.
```

Filename format: `YYYY-MM-DD-slug.md` (e.g., `2025-02-23-pensieve-idea.md`)

## Building

```bash
# Install dependencies
go mod tidy

# Run the generator
go run main.go

# Or build a binary
go build -o pensieve main.go
./pensieve
```

## Output

The generator creates:

- `output/index.html` - Latest 20 thoughts
- `output/all.html` - All thoughts (single page)
- `output/thoughts/*.html` - Individual thought pages
- `output/tags/*.html` - Tag pages
- `output/archive/` - Archive by month
  - `index.html` - List of all months
  - `YYYY-MM.html` - Individual month pages

## Features

✅ Latest 20 on homepage  
✅ Archive by month  
✅ Tag pages  
✅ All thoughts on one page (Ctrl+F friendly)  
✅ Prev/next navigation on thought pages  
✅ Reverse chronological order  
✅ Pure HTML output, no JavaScript needed  
✅ Fast: handles 3000+ thoughts in <1 second  

## GitHub Actions

Add this to `.github/workflows/build.yml`:

```yaml
name: Build and Deploy

on:
  push:
    branches: [main]
    paths:
      - 'thoughts/**'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build site
        run: |
          go mod tidy
          go run main.go
      
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./output
```

## Deploy

The `output/` directory is your complete static site. Deploy it anywhere:

- **GitHub Pages**: Enable in repo settings, push to `gh-pages` branch
- **Netlify**: Drag and drop the `output/` folder
- **Vercel**: Connect repo, set output directory to `output/`
- **Any static host**: Upload `output/` contents

## Philosophy

- Pure HTML, minimal CSS
- No JavaScript required
- No database, no backend
- Git is your CMS
- One binary, no dependencies in production
- Scales to thousands of thoughts
- Fully searchable with browser Ctrl+F
