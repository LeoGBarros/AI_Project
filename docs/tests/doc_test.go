package docs_tests

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ============================================================
// Helpers
// ============================================================

const projectRoot = "../../"

// excludedPrefixes lists directories whose Markdown files should NOT be tested.
var excludedPrefixes = []string{
	filepath.FromSlash(".kiro/specs/"),
	filepath.FromSlash(".kiro/"),
	filepath.FromSlash(".cursor/"),
	filepath.FromSlash(".git/"),
}

// collectMarkdownFiles walks the project root and returns all .md files,
// excluding directories listed in excludedPrefixes.
func collectMarkdownFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	root := projectRoot
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		for _, prefix := range excludedPrefixes {
			if strings.HasPrefix(filepath.ToSlash(rel), filepath.ToSlash(prefix)) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk project root: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no markdown files found")
	}
	return files
}

// readFileContent reads the full content of a file.
func readFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// relPath returns a human-readable relative path from the project root.
func relPath(path string) string {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return path
	}
	return rel
}

// mdFileGen creates a gopter generator that yields indices into a slice of markdown files.
func mdFileGen(files []string) gopter.Gen {
	return func(params *gopter.GenParameters) *gopter.GenResult {
		idx := int(params.NextUint64() % uint64(len(files)))
		return gopter.NewGenResult(idx, gopter.NoShrinker)
	}
}

// parseCodeBlocks extracts code blocks from markdown content.
// Returns a slice of codeBlock structs.
type codeBlock struct {
	lang    string // language tag (empty if bare ```)
	content string
	line    int // 1-based line number of the opening ```
}

func parseCodeBlocks(content string) []codeBlock {
	var blocks []codeBlock
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	inBlock := false
	var current codeBlock
	var builder strings.Builder

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inBlock && strings.HasPrefix(trimmed, "```") {
			inBlock = true
			current = codeBlock{
				lang: strings.TrimPrefix(trimmed, "```"),
				line: lineNum,
			}
			builder.Reset()
			continue
		}
		if inBlock && strings.TrimSpace(line) == "```" {
			current.content = builder.String()
			blocks = append(blocks, current)
			inBlock = false
			continue
		}
		if inBlock {
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}
	return blocks
}

// parseHeadings extracts headings from markdown content.
type heading struct {
	level int
	text  string
	line  int
}

func parseHeadings(content string) []heading {
	var headings []heading
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	inCodeBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			if level >= 1 && level <= 6 {
				text := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
				headings = append(headings, heading{level: level, text: text, line: lineNum})
			}
		}
	}
	return headings
}

// linkRegex matches Markdown links: [text](url)
var linkRegex = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

// extractInternalLinks extracts internal links from markdown content.
// Internal links are those that don't start with http://, https://, or mailto:
type mdLink struct {
	text   string
	target string
	line   int
}

func extractInternalLinks(content string) []mdLink {
	var links []mdLink
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	inCodeBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		matches := linkRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			target := m[2]
			// Skip external links and anchors-only
			if strings.HasPrefix(target, "http://") ||
				strings.HasPrefix(target, "https://") ||
				strings.HasPrefix(target, "mailto:") ||
				strings.HasPrefix(target, "#") {
				continue
			}
			links = append(links, mdLink{text: m[1], target: target, line: lineNum})
		}
	}
	return links
}

// ============================================================
// Property 1: Representações visuais usam blocos de código apropriados
// Feature: project-documentation-improvement, Property 1: Representações visuais usam blocos de código apropriados
// **Validates: Requirements 1.1, 1.5**
// ============================================================

// boxDrawingChars are characters used in ASCII diagrams that must be inside ```text blocks.
var boxDrawingChars = []rune{'┌', '─', '┐', '│', '└', '┘', '├', '┤', '┬', '┴', '┼', '▼', '▲', '►', '◄', '═', '╔', '╗', '╚', '╝', '║'}

// mermaidKeywords are keywords that indicate Mermaid diagram content.
var mermaidKeywords = []string{"sequenceDiagram", "graph ", "graph\n", "flowchart ", "flowchart\n", "stateDiagram", "classDiagram", "erDiagram", "gantt", "pie ", "pie\n"}

func containsBoxDrawing(line string) bool {
	for _, ch := range line {
		for _, bd := range boxDrawingChars {
			if ch == bd {
				return true
			}
		}
	}
	return false
}

func containsMermaidKeyword(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, kw := range mermaidKeywords {
		if strings.HasPrefix(trimmed, strings.TrimSpace(kw)) {
			return true
		}
	}
	return false
}

func TestProperty1_VisualRepresentationsInCodeBlocks(t *testing.T) {
	files := collectMarkdownFiles(t)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = len(files)
	parameters.MaxSize = len(files)
	properties := gopter.NewProperties(parameters)

	properties.Property("visual representations use appropriate code blocks", prop.ForAll(
		func(idx int) bool {
			path := files[idx]
			content, err := readFileContent(path)
			if err != nil {
				t.Errorf("failed to read %s: %v", relPath(path), err)
				return false
			}

			// Parse lines outside of code blocks
			scanner := bufio.NewScanner(strings.NewReader(content))
			lineNum := 0
			inCodeBlock := false
			ok := true

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				trimmed := strings.TrimSpace(line)

				if strings.HasPrefix(trimmed, "```") {
					inCodeBlock = !inCodeBlock
					continue
				}

				if !inCodeBlock {
					if containsBoxDrawing(line) {
						t.Errorf("%s:%d — box-drawing characters found outside code block: %s",
							relPath(path), lineNum, strings.TrimSpace(line))
						ok = false
					}
					if containsMermaidKeyword(line) {
						t.Errorf("%s:%d — Mermaid keyword found outside code block: %s",
							relPath(path), lineNum, strings.TrimSpace(line))
						ok = false
					}
				}
			}
			return ok
		},
		mdFileGen(files),
	))

	properties.TestingRun(t)
}

// ============================================================
// Property 2: Hierarquia de headings é consistente
// Feature: project-documentation-improvement, Property 2: Hierarquia de headings é consistente
// **Validates: Requirements 1.4**
// ============================================================

func TestProperty2_HeadingHierarchyConsistent(t *testing.T) {
	files := collectMarkdownFiles(t)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = len(files)
	parameters.MaxSize = len(files)
	properties := gopter.NewProperties(parameters)

	properties.Property("heading hierarchy is consistent", prop.ForAll(
		func(idx int) bool {
			path := files[idx]
			content, err := readFileContent(path)
			if err != nil {
				t.Errorf("failed to read %s: %v", relPath(path), err)
				return false
			}

			headings := parseHeadings(content)
			if len(headings) == 0 {
				return true // no headings, nothing to check
			}

			ok := true

			// h1 appears at most once
			h1Count := 0
			for _, h := range headings {
				if h.level == 1 {
					h1Count++
				}
			}
			if h1Count > 1 {
				t.Errorf("%s — h1 appears %d times (expected at most 1)", relPath(path), h1Count)
				ok = false
			}

			// No level skips: each heading level must not jump more than 1 from the previous
			// e.g., h1 → h3 without h2 is invalid
			prevLevel := 0
			for _, h := range headings {
				if prevLevel > 0 && h.level > prevLevel+1 {
					t.Errorf("%s:%d — heading level skip: h%d → h%d (heading: %q)",
						relPath(path), h.line, prevLevel, h.level, h.text)
					ok = false
				}
				prevLevel = h.level
			}

			return ok
		},
		mdFileGen(files),
	))

	properties.TestingRun(t)
}

// ============================================================
// Property 3: Blocos de código possuem syntax highlighting
// Feature: project-documentation-improvement, Property 3: Blocos de código possuem syntax highlighting
// **Validates: Requirements 1.6**
// ============================================================

func TestProperty3_CodeBlocksHaveSyntaxHighlighting(t *testing.T) {
	files := collectMarkdownFiles(t)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = len(files)
	parameters.MaxSize = len(files)
	properties := gopter.NewProperties(parameters)

	properties.Property("code blocks have syntax highlighting tags", prop.ForAll(
		func(idx int) bool {
			path := files[idx]
			content, err := readFileContent(path)
			if err != nil {
				t.Errorf("failed to read %s: %v", relPath(path), err)
				return false
			}

			blocks := parseCodeBlocks(content)
			ok := true

			for _, block := range blocks {
				if block.lang == "" {
					// Bare ``` with no language tag — this is the violation
					t.Errorf("%s:%d — code block without language tag (bare ```)",
						relPath(path), block.line)
					ok = false
				}
			}

			return ok
		},
		mdFileGen(files),
	))

	properties.TestingRun(t)
}

// ============================================================
// Property 4: Links internos usam caminhos relativos
// Feature: project-documentation-improvement, Property 4: Links internos usam caminhos relativos
// **Validates: Requirements 1.9**
// ============================================================

func TestProperty4_InternalLinksUseRelativePaths(t *testing.T) {
	files := collectMarkdownFiles(t)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = len(files)
	parameters.MaxSize = len(files)
	properties := gopter.NewProperties(parameters)

	properties.Property("internal links use relative paths", prop.ForAll(
		func(idx int) bool {
			path := files[idx]
			content, err := readFileContent(path)
			if err != nil {
				t.Errorf("failed to read %s: %v", relPath(path), err)
				return false
			}

			links := extractInternalLinks(content)
			ok := true

			for _, link := range links {
				// Internal links should not start with / (absolute filesystem path)
				if strings.HasPrefix(link.target, "/") {
					t.Errorf("%s:%d — internal link uses absolute path: [%s](%s)",
						relPath(path), link.line, link.text, link.target)
					ok = false
				}
			}

			return ok
		},
		mdFileGen(files),
	))

	properties.TestingRun(t)
}

// ============================================================
// Property 5: Links internos resolvem para arquivos existentes
// Feature: project-documentation-improvement, Property 5: Links internos resolvem para arquivos existentes
// **Validates: Requirements 8.2, 9.1**
// ============================================================

func TestProperty5_InternalLinksResolveToExistingFiles(t *testing.T) {
	files := collectMarkdownFiles(t)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = len(files)
	parameters.MaxSize = len(files)
	properties := gopter.NewProperties(parameters)

	properties.Property("internal links resolve to existing files", prop.ForAll(
		func(idx int) bool {
			path := files[idx]
			content, err := readFileContent(path)
			if err != nil {
				t.Errorf("failed to read %s: %v", relPath(path), err)
				return false
			}

			links := extractInternalLinks(content)
			sourceDir := filepath.Dir(path)
			ok := true

			for _, link := range links {
				target := link.target

				// Strip anchor fragments (#section)
				if anchorIdx := strings.Index(target, "#"); anchorIdx >= 0 {
					target = target[:anchorIdx]
				}

				// Skip empty targets (pure anchor links that were already filtered)
				if target == "" {
					continue
				}

				// Resolve relative to the source file's directory
				resolved := filepath.Join(sourceDir, target)
				resolved = filepath.Clean(resolved)

				if _, err := os.Stat(resolved); os.IsNotExist(err) {
					t.Errorf("%s:%d — broken link: [%s](%s) → %s does not exist",
						relPath(path), link.line, link.text, link.target, relPath(resolved))
					ok = false
				}
			}

			return ok
		},
		mdFileGen(files),
	))

	properties.TestingRun(t)
}

// ============================================================
// Property 6: Nomes de arquivos de diagramas contêm o nome do pattern
// Feature: project-documentation-improvement, Property 6: Nomes de arquivos de diagramas contêm o nome do pattern
// **Validates: Requirements 4.5, 8.1**
// ============================================================

var validPatternNames = []string{
	"hexagonal",
	"pkce",
	"ropc",
	"client-credentials",
	"circuit-breaker",
	"pubsub",
	"token-refresh",
}

func TestProperty6_DiagramFilenamesContainPatternName(t *testing.T) {
	diagramsDir := filepath.Join(projectRoot, "docs", "diagrams")

	entries, err := os.ReadDir(diagramsDir)
	if err != nil {
		t.Fatalf("failed to read diagrams directory: %v", err)
	}

	var diagramFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		// Exclude README.md
		if strings.EqualFold(name, "README.md") {
			continue
		}
		diagramFiles = append(diagramFiles, name)
	}

	if len(diagramFiles) == 0 {
		t.Fatal("no diagram files found in docs/diagrams/")
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = len(diagramFiles)
	parameters.MaxSize = len(diagramFiles)
	properties := gopter.NewProperties(parameters)

	properties.Property("diagram filenames contain a pattern name", prop.ForAll(
		func(idx int) bool {
			filename := diagramFiles[idx]
			lower := strings.ToLower(filename)

			for _, pattern := range validPatternNames {
				if strings.Contains(lower, pattern) {
					return true
				}
			}

			t.Errorf("docs/diagrams/%s — filename does not contain any valid pattern name %v",
				filename, validPatternNames)
			return false
		},
		func(params *gopter.GenParameters) *gopter.GenResult {
			idx := int(params.NextUint64() % uint64(len(diagramFiles)))
			return gopter.NewGenResult(idx, gopter.NoShrinker)
		},
	))

	properties.TestingRun(t)
}
