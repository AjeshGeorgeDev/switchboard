package security

import (
	"context"
	"time"

	"github.com/switchboard/switchboard/internal/db"
)

type OverviewStats struct {
	CriticalCount       int64 `json:"critical_count"`
	HighCount           int64 `json:"high_count"`
	MediumCount         int64 `json:"medium_count"`
	LowCount            int64 `json:"low_count"`
	FixableCritical     int64 `json:"fixable_critical"`
	FixableCriticalHigh int64 `json:"fixable_critical_high"`
	UnfixedCriticalHigh int64 `json:"unfixed_critical_high"`
	NewThisWeek         int64 `json:"new_this_week"`
	AgingLt7d           int64 `json:"aging_lt_7d"`
	Aging7To30d         int64 `json:"aging_7_to_30d"`
	AgingGt30d          int64 `json:"aging_gt_30d"`
}

type RiskyImage struct {
	ImageName          string     `json:"image_name"`
	LatestTag          string     `json:"latest_tag"`
	TagCount           int64      `json:"tag_count,omitempty"`
	CriticalCount      int64      `json:"critical_count"`
	HighCount          int64      `json:"high_count"`
	MediumCount        int64      `json:"medium_count"`
	LowCount           int64      `json:"low_count"`
	TotalCount         int64      `json:"total_count"`
	OldestCriticalAt   *time.Time `json:"oldest_critical_at,omitempty"`
	OldestCriticalDays *int       `json:"oldest_critical_days,omitempty"`
}

type Overview struct {
	Stats      OverviewStats `json:"stats"`
	TopImages  []RiskyImage  `json:"top_images"`
	GeneratedAt time.Time    `json:"generated_at"`
}

func BuildOverview(ctx context.Context, q *db.Queries, topN int32) (Overview, error) {
	if topN <= 0 {
		topN = 10
	}
	stats, err := q.GetCVEOverviewStats(ctx)
	if err != nil {
		return Overview{}, err
	}
	rows, err := q.ListTopRiskyImages(ctx, topN)
	if err != nil {
		return Overview{}, err
	}
	images := make([]RiskyImage, 0, len(rows))
	for _, row := range rows {
		images = append(images, mapTopImage(row))
	}
	return Overview{
		Stats: OverviewStats{
			CriticalCount:       stats.CriticalCount,
			HighCount:           stats.HighCount,
			MediumCount:         stats.MediumCount,
			LowCount:            stats.LowCount,
			FixableCritical:     stats.FixableCritical,
			FixableCriticalHigh: stats.FixableCriticalHigh,
			UnfixedCriticalHigh: stats.UnfixedCriticalHigh,
			NewThisWeek:         stats.NewThisWeek,
			AgingLt7d:           stats.AgingLt7d,
			Aging7To30d:         stats.Aging7To30d,
			AgingGt30d:          stats.AgingGt30d,
		},
		TopImages:   images,
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func mapTopImage(row db.ListTopRiskyImagesRow) RiskyImage {
	img := RiskyImage{
		ImageName:     row.ImageName,
		LatestTag:     row.LatestTag,
		CriticalCount: row.CriticalCount,
		HighCount:     row.HighCount,
		MediumCount:   row.MediumCount,
		LowCount:      row.LowCount,
		TotalCount:    row.TotalCount,
	}
	if t, ok := asTime(row.OldestCriticalAt); ok {
		img.OldestCriticalAt = &t
		days := int(time.Since(t).Hours() / 24)
		if days < 0 {
			days = 0
		}
		img.OldestCriticalDays = &days
	}
	return img
}

func mapRollupImage(row db.ListImageRiskRollupRow) RiskyImage {
	img := RiskyImage{
		ImageName:     row.ImageName,
		LatestTag:     row.LatestTag,
		TagCount:      row.TagCount,
		CriticalCount: row.CriticalCount,
		HighCount:     row.HighCount,
		MediumCount:   row.MediumCount,
		LowCount:      row.LowCount,
		TotalCount:    row.TotalCount,
	}
	if t, ok := asTime(row.OldestCriticalAt); ok {
		img.OldestCriticalAt = &t
		days := int(time.Since(t).Hours() / 24)
		if days < 0 {
			days = 0
		}
		img.OldestCriticalDays = &days
	}
	return img
}

func asTime(v interface{}) (time.Time, bool) {
	if v == nil {
		return time.Time{}, false
	}
	switch t := v.(type) {
	case time.Time:
		return t, true
	case *time.Time:
		if t == nil {
			return time.Time{}, false
		}
		return *t, true
	default:
		return time.Time{}, false
	}
}

func findingAgeDays(createdAt time.Time) int {
	if createdAt.IsZero() {
		return 0
	}
	days := int(time.Since(createdAt).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func isFixable(fixedVersion string) bool {
	return len(fixedVersion) > 0
}
