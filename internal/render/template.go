package render

// boardTemplate is the self-contained HTML template for the Kanban board.
// It expects a *board.Board as its data.
const boardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Markban</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #f0f2f5;
      color: #2c3e50;
      min-height: 100vh;
      padding: 2rem;
    }

    h1 {
      font-size: 1.4rem;
      font-weight: 600;
      margin-bottom: 1.5rem;
      color: #4a5568;
      letter-spacing: 0.02em;
    }

    .board {
      display: flex;
      flex-direction: row;
      gap: 1.25rem;
      align-items: flex-start;
      overflow-x: auto;
    }

    .column {
      background: #e2e8f0;
      border-radius: 10px;
      padding: 1rem;
      min-width: 260px;
      max-width: 320px;
      flex: 0 0 auto;
    }

    .column-header {
      font-size: 0.85rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: #718096;
      margin-bottom: 0.875rem;
      padding-bottom: 0.5rem;
      border-bottom: 2px solid #cbd5e0;
    }

    .cards {
      display: flex;
      flex-direction: column;
      gap: 0.625rem;
    }

    .card {
      background: #ffffff;
      border-radius: 7px;
      padding: 0.75rem 1rem;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.04);
      transition: box-shadow 0.15s ease;
    }

    .card:hover {
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.10), 0 2px 4px rgba(0, 0, 0, 0.06);
    }

    .card-title {
      font-size: 0.9rem;
      font-weight: 600;
      color: #2d3748;
      line-height: 1.4;
    }

    .card-blurb {
      font-size: 0.8rem;
      color: #718096;
      margin-top: 0.3rem;
      line-height: 1.5;
    }

    .card-footer {
      display: flex;
      justify-content: space-between;
      margin-top: 0.5rem;
    }

    .card-slug {
      font-size: 0.7rem;
      font-style: italic;
      color: #a0aec0;
    }

    .card-priority {
      font-size: 0.7rem;
      font-style: italic;
      color: #a0aec0;
    }

    .empty {
      font-size: 0.8rem;
      color: #a0aec0;
      font-style: italic;
      text-align: center;
      padding: 0.5rem 0;
    }

    .archive-link {
      position: fixed;
      bottom: 1.5rem;
      right: 2rem;
      font-size: 0.85rem;
      color: #718096;
      text-decoration: none;
      transition: color 0.15s ease;
    }

    .archive-link:hover {
      color: #2d3748;
      text-decoration: underline;
    }

    /* Colored cards always use dark text for readability */
    .card-colored .card-title {
      color: #2d3748;
    }

    .card-colored .card-blurb {
      color: #4a5568;
    }

    .card-colored .card-slug,
    .card-colored .card-priority {
      color: #718096;
    }

    @media (prefers-color-scheme: dark) {
      body {
        background: #1a202c;
        color: #e2e8f0;
      }

      h1 {
        color: #a0aec0;
      }

      .column {
        background: #2d3748;
      }

      .column-header {
        color: #a0aec0;
        border-bottom-color: #4a5568;
      }

      .card {
        background: #2d3748;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3), 0 1px 2px rgba(0, 0, 0, 0.2);
      }

      .card:hover {
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.4), 0 2px 4px rgba(0, 0, 0, 0.3);
      }

      .card-title {
        color: #e2e8f0;
      }

      .card-blurb {
        color: #a0aec0;
      }

      .card-slug, .card-priority {
        color: #718096;
      }

      .empty {
        color: #718096;
      }

      .archive-link {
        color: #a0aec0;
      }

      .archive-link:hover {
        color: #e2e8f0;
      }

      /* Colored cards keep dark text even in dark mode */
      .card-colored .card-title {
        color: #2d3748;
      }

      .card-colored .card-blurb {
        color: #4a5568;
      }

      .card-colored .card-slug,
      .card-colored .card-priority {
        color: #718096;
      }
    }
  </style>
</head>
<body>
  <h1>Markban{{if .Board.Name}} - {{.Board.Name}}{{end}}</h1>
  <div class="board">
    {{- range .Board.Columns}}
    <div class="column">
      <div class="column-header">{{.Name}}</div>
      <div class="cards">
        {{- if .Tasks}}
         {{- range .Tasks}}
         <div class="card{{if .Color}} card-colored{{end}}" style="{{cardStyle .Color}}">
           <div class="card-title">{{.Title}}</div>
           {{- if .Blurb}}
           <div class="card-blurb">{{.Blurb}}</div>
           {{- end}}
            <div class="card-footer">
              {{- if not .Done.IsZero}}
              <div class="card-priority">[{{.Done.Format "2006-01-02"}}]</div>
              {{- else if .Priority}}
              <div class="card-priority">[{{.Priority}}]</div>
              {{- end}}
              <div class="card-slug">[{{.Slug}}]</div>
            </div>
         </div>
         {{- end}}
        {{- else}}
        <div class="empty">No tasks</div>
        {{- end}}
      </div>
    </div>
    {{- end}}
  </div>
  {{- if .HasArchive}}
  <a href="/archive" class="archive-link">Archive</a>
  {{- end}}
  <script>
    const es = new EventSource("/events");
    es.addEventListener("reload", () => location.reload());
  </script>
</body>
</html>
`

// archiveTemplate is the HTML template for the archive page.
// It expects a []board.Task as its data.
const archiveTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Markban - Archive</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #f0f2f5;
      color: #2c3e50;
      min-height: 100vh;
      padding: 2rem;
    }

    h1 {
      font-size: 1.4rem;
      font-weight: 600;
      margin-bottom: 1.5rem;
      color: #4a5568;
      letter-spacing: 0.02em;
    }

    .archive-list {
      max-width: 800px;
    }

    .archive-item {
      background: #ffffff;
      border-radius: 7px;
      padding: 0.75rem 1rem;
      margin-bottom: 0.625rem;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.04);
    }

    .archive-title {
      font-size: 0.9rem;
      font-weight: 600;
      color: #2d3748;
      line-height: 1.4;
    }

    .archive-blurb {
      font-size: 0.8rem;
      color: #718096;
      margin-top: 0.3rem;
      line-height: 1.5;
    }

    .empty {
      font-size: 0.8rem;
      color: #a0aec0;
      font-style: italic;
    }

    .back-link {
      display: inline-block;
      margin-bottom: 1.5rem;
      font-size: 0.85rem;
      color: #718096;
      text-decoration: none;
    }

    .back-link:hover {
      color: #2d3748;
      text-decoration: underline;
    }

    @media (prefers-color-scheme: dark) {
      body {
        background: #1a202c;
        color: #e2e8f0;
      }

      h1 {
        color: #a0aec0;
      }

      .archive-item {
        background: #2d3748;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3), 0 1px 2px rgba(0, 0, 0, 0.2);
      }

      .archive-title {
        color: #e2e8f0;
      }

      .archive-blurb {
        color: #a0aec0;
      }

      .empty {
        color: #718096;
      }

      .back-link {
        color: #a0aec0;
      }

      .back-link:hover {
        color: #e2e8f0;
      }
    }
  </style>
</head>
<body>
  <a href="/" class="back-link">&larr; Back to board</a>
  <h1>Archive</h1>
  <div class="archive-list">
    {{- if .Tasks}}
    {{- range .Tasks}}
    <div class="archive-item">
      <div class="archive-title">{{.Title}}</div>
      {{- if .Blurb}}
      <div class="archive-blurb">{{.Blurb}}</div>
      {{- end}}
    </div>
    {{- end}}
    {{- else}}
    <div class="empty">No archived tasks</div>
    {{- end}}
  </div>
</body>
</html>
`
