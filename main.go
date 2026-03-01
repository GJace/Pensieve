package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

type Thought struct {
	Filename    string
	Date        time.Time
	DateString  string
	Title       string
	Content     template.HTML
	Preview     template.HTML
	Tags        []string
	Slug        string
	HTMLPath    string
	PrevThought *Thought
	NextThought *Thought
	IsLong      bool
}

type FrontMatter struct {
	Tags string `yaml:"tags"`
	Date string `yaml:"date"`
	Time string `yaml:"time"`
}

type PageData struct {
	Title      string
	Subtitle   string
	Thoughts   []Thought
	AllTags    []string
	Tag        string
	PrevMonth  string
	NextMonth  string
	MonthTitle string
}

type ArchiveMonth struct {
	YearMonth string
	Title     string
	Count     int
}

type ArchiveYear struct {
	Year   string
	Months []ArchiveMonth
}

func main() {
	thoughtsDir := "thoughts"
	outputDir := "output"
	templatesDir := "templates"

	// Clean and create output directory
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir+"/thoughts", 0755)
	os.MkdirAll(outputDir+"/archive", 0755)
	os.MkdirAll(outputDir+"/tags", 0755)

	// Copy CSS
	copyFile("style.css", outputDir+"/style.css")

	// Read all thoughts
	thoughts, err := readThoughts(thoughtsDir)
	if err != nil {
		fmt.Printf("Error reading thoughts: %v\n", err)
		return
	}

	// Sort by date (newest first)
	sort.Slice(thoughts, func(i, j int) bool {
		return thoughts[i].Date.After(thoughts[j].Date)
	})

	// Set prev/next links
	for i := range thoughts {
		if i > 0 {
			thoughts[i].NextThought = &thoughts[i-1]
		}
		if i < len(thoughts)-1 {
			thoughts[i].PrevThought = &thoughts[i+1]
		}
	}

	// Generate individual thought pages
	for _, thought := range thoughts {
		generateThoughtPage(thought, templatesDir, outputDir)
	}

	// Collect all tags
	tagMap := make(map[string][]Thought)
	for _, thought := range thoughts {
		for _, tag := range thought.Tags {
			tagMap[tag] = append(tagMap[tag], thought)
		}
	}

	allTags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		allTags = append(allTags, tag)
	}
	sort.Strings(allTags)

	// Generate index (latest 20)
	latest := thoughts
	if len(latest) > 20 {
		latest = thoughts[:20]
	}
	generateIndexPage(latest, allTags, templatesDir, outputDir)

	// Generate all.html
	generateAllPage(thoughts, allTags, templatesDir, outputDir)

	// Generate tag pages
	for tag, taggedThoughts := range tagMap {
		generateTagPage(tag, taggedThoughts, allTags, templatesDir, outputDir)
	}

	// Generate archive pages
	generateArchivePages(thoughts, templatesDir, outputDir)

	fmt.Printf("Generated %d thoughts\n", len(thoughts))
	fmt.Printf("Generated %d tag pages\n", len(allTags))
}

func readThoughts(dir string) ([]Thought, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var thoughts []Thought
	md := goldmark.New(goldmark.WithExtensions())

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			continue
		}

		// Parse frontmatter and content
		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			continue
		}

		var fm FrontMatter
		yaml.Unmarshal([]byte(parts[1]), &fm)

		// Parse date and time
		var date time.Time
		if fm.Time != "" {
			// Parse with time if provided
			dateTimeStr := fm.Date + " " + fm.Time
			date, _ = time.Parse("2006-01-02 15:04", dateTimeStr)
		} else {
			// Default to 00:00 if no time provided
			date, _ = time.Parse("2006-01-02", fm.Date)
		}

		// Convert markdown to HTML
		var buf strings.Builder
		md.Convert([]byte(parts[2]), &buf, parser.WithContext(parser.NewContext()))

		// Full content
		htmlContent := buf.String()
		
		// Create preview
		preview := htmlContent
		plainText := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(htmlContent, "<p>", ""), "</p>", ""))
		isLong := len(plainText) > 200
		
		if isLong {
			// Truncate at ~200 chars, at sentence boundary
			truncated := plainText
			if len(plainText) > 200 {
				truncated = plainText[:200]
				// Find last sentence ending (. ! ?)
				lastPeriod := strings.LastIndexAny(truncated, ".!?")
				if lastPeriod > 0 {
					truncated = plainText[:lastPeriod+1]
				}
			}
			preview = "<p>" + truncated + "</p>"
		}

		// Create slug from filename
		slug := strings.TrimSuffix(file.Name(), ".md")

		thought := Thought{
			Filename:   file.Name(),
			Date:       date,
			DateString: date.Format("2006-01-02"),
			Content:    template.HTML(htmlContent),
			Preview:    template.HTML(preview),
			Tags:       strings.Split(strings.TrimSpace(fm.Tags), ","),
			Slug:       slug,
			HTMLPath:   "thoughts/" + slug + ".html",
			IsLong:     isLong,
		}

		// Trim spaces from tags
		for i, tag := range thought.Tags {
			thought.Tags[i] = strings.TrimSpace(tag)
		}

		thoughts = append(thoughts, thought)
	}

	return thoughts, nil
}

func generateThoughtPage(thought Thought, templatesDir, outputDir string) {
	tmpl := template.Must(template.ParseFiles(filepath.Join(templatesDir, "thought.html")))

	data := PageData{
		Thoughts: []Thought{thought},
	}

	f, _ := os.Create(filepath.Join(outputDir, thought.HTMLPath))
	defer f.Close()

	tmpl.Execute(f, data)
}

func generateIndexPage(thoughts []Thought, allTags []string, templatesDir, outputDir string) {
	tmpl := template.Must(template.ParseFiles(filepath.Join(templatesDir, "index.html")))

	data := PageData{
		Title:    "Pensieve",
		Subtitle: "Latest thoughts",
		Thoughts: thoughts,
		AllTags:  allTags,
	}

	f, _ := os.Create(filepath.Join(outputDir, "index.html"))
	defer f.Close()

	tmpl.Execute(f, data)
}

func generateAllPage(thoughts []Thought, allTags []string, templatesDir, outputDir string) {
	tmpl := template.Must(template.ParseFiles(filepath.Join(templatesDir, "all.html")))

	data := PageData{
		Title:    "All Thoughts",
		Subtitle: fmt.Sprintf("All thoughts (%d total)", len(thoughts)),
		Thoughts: thoughts,
		AllTags:  allTags,
	}

	f, _ := os.Create(filepath.Join(outputDir, "all.html"))
	defer f.Close()

	tmpl.Execute(f, data)
}

func generateTagPage(tag string, thoughts []Thought, allTags []string, templatesDir, outputDir string) {
	tmpl := template.Must(template.ParseFiles(filepath.Join(templatesDir, "tag.html")))

	data := PageData{
		Title:    "Tag: " + tag,
		Tag:      tag,
		Thoughts: thoughts,
		AllTags:  allTags,
	}

	f, _ := os.Create(filepath.Join(outputDir, "tags", tag+".html"))
	defer f.Close()

	tmpl.Execute(f, data)
}

func generateArchivePages(thoughts []Thought, templatesDir, outputDir string) {
	// Group by month
	monthMap := make(map[string][]Thought)
	for _, thought := range thoughts {
		monthKey := thought.Date.Format("2006-01")
		monthMap[monthKey] = append(monthMap[monthKey], thought)
	}

	// Create archive index
	var months []string
	for month := range monthMap {
		months = append(months, month)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(months)))

	// Group by year for archive index
	yearMap := make(map[string][]ArchiveMonth)
	for _, month := range months {
		year := month[:4]
		t, _ := time.Parse("2006-01", month)
		yearMap[year] = append(yearMap[year], ArchiveMonth{
			YearMonth: month,
			Title:     t.Format("January 2006"),
			Count:     len(monthMap[month]),
		})
	}

	var years []string
	for year := range yearMap {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	var archiveYears []ArchiveYear
	for _, year := range years {
		archiveYears = append(archiveYears, ArchiveYear{
			Year:   year,
			Months: yearMap[year],
		})
	}

	// Generate archive index
	tmpl := template.Must(template.ParseFiles(filepath.Join(templatesDir, "archive-index.html")))
	f, _ := os.Create(filepath.Join(outputDir, "archive", "index.html"))
	defer f.Close()
	tmpl.Execute(f, map[string]interface{}{"Years": archiveYears})

	// Generate individual month pages
	for i, month := range months {
		t, _ := time.Parse("2006-01", month)
		monthTitle := t.Format("January 2006")

		var prevMonth, nextMonth string
		if i < len(months)-1 {
			prevMonth = months[i+1]
		}
		if i > 0 {
			nextMonth = months[i-1]
		}

		generateMonthPage(month, monthTitle, monthMap[month], prevMonth, nextMonth, templatesDir, outputDir)
	}
}

func generateMonthPage(month, monthTitle string, thoughts []Thought, prevMonth, nextMonth, templatesDir, outputDir string) {
	tmpl := template.Must(template.ParseFiles(filepath.Join(templatesDir, "archive-month.html")))

	data := PageData{
		Title:      monthTitle,
		MonthTitle: monthTitle,
		Thoughts:   thoughts,
		PrevMonth:  prevMonth,
		NextMonth:  nextMonth,
	}

	f, _ := os.Create(filepath.Join(outputDir, "archive", month+".html"))
	defer f.Close()

	tmpl.Execute(f, data)
}

func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
