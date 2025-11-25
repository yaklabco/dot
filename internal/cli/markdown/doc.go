// Package markdown provides terminal-friendly Markdown rendering.
//
// The renderer converts common Markdown elements to styled terminal output
// using ANSI escape codes via the render.Colorizer.
//
// Supported elements:
//   - Headers (##, ###) rendered as bold
//   - Bullet lists (-, *) rendered with styled markers
//   - Bold text (**text**) rendered as bold
//   - Inline code (`code`) rendered as accent color
//   - Links [text](url) rendered as text only
package markdown
