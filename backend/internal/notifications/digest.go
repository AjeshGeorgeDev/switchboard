package notifications

import (
	"bytes"
	"embed"
	"html/template"
	"strconv"
	"strings"
)

//go:embed templates/weekly_digest.html
var digestTemplates embed.FS

type DigestImageRow struct {
	Name       string
	Critical   int64
	High       int64
	OldestDays *int
}

type DigestEmailData struct {
	Title           string
	Intro           string
	OverviewURL     string
	Critical        int64
	High            int64
	NewThisWeek     int64
	FixableCritical int64
	AgingLt7d       int64
	AgingGt30d      int64
	TopImages       []DigestImageRow
}

func RenderWeeklyDigestHTML(data DigestEmailData) (string, error) {
	tmpl, err := template.ParseFS(digestTemplates, "templates/weekly_digest.html")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func PlainDigestBody(data DigestEmailData) string {
	var b strings.Builder
	b.WriteString(data.Intro)
	b.WriteString("\n\n")
	b.WriteString("Critical: ")
	b.WriteString(strconv.FormatInt(data.Critical, 10))
	b.WriteString(" | High: ")
	b.WriteString(strconv.FormatInt(data.High, 10))
	b.WriteString(" | New this week: ")
	b.WriteString(strconv.FormatInt(data.NewThisWeek, 10))
	b.WriteString(" | Fixable criticals: ")
	b.WriteString(strconv.FormatInt(data.FixableCritical, 10))
	if len(data.TopImages) > 0 {
		b.WriteString("\n\nRiskiest images:\n")
		for _, img := range data.TopImages {
			b.WriteString("- ")
			b.WriteString(img.Name)
			b.WriteString(": critical ")
			b.WriteString(strconv.FormatInt(img.Critical, 10))
			b.WriteString(", high ")
			b.WriteString(strconv.FormatInt(img.High, 10))
			b.WriteString("\n")
		}
	}
	if data.OverviewURL != "" {
		b.WriteString("\n")
		b.WriteString(data.OverviewURL)
	}
	return b.String()
}
