package domain

import (
	"fmt"
	"strings"
	"time"
)

type field struct {
	name  string
	value string
}

// Универсальная функция для любых указателей
func ptrStr[T any](p *T) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *p)
}

// Специальная функция для Location
func geoStr(g *Location) string {
	if g == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Location{Lat: %f, Lon: %f}", g.Lat, g.Lon)
}

func timeStr(t *time.Time) string {
	if t == nil {
		return "<nil>"
	}
	return t.Format(time.RFC3339)
}

func (u *User) String() string {
	if u == nil {
		return "<nil>"
	}
	return structString("User", []field{
		{"ID", ptrStr(u.ID)},
		{"Login", ptrStr(u.Login)},
		{"Username", ptrStr(u.Username)},
		{"Password", ifStr(u.Password != nil, "[hidden]")},
		{"Description", ptrStr(u.Description)},
		{"Comment", ptrStr(u.Comment)},
		{"RegDate", timeStr(u.RegDate)},
		{"Location", geoStr(u.Location)},
		{"SocialNet", ptrStr(u.SocialNet)},
	})
}

func (f *UserFilter) String() string {
	if f == nil {
		return "<nil>"
	}
	return structString("UserFilter", []field{
		{"Search", ptrStr(f.Search)},
		{"DateFrom", timeStr(f.DateFrom)},
		{"DateTo", timeStr(f.DateTo)},
		{"Lat", ptrStr(f.Lat)},
		{"Lon", ptrStr(f.Lon)},
		{"Distance", ptrStr(f.Distance)},
		{"SocialType", ptrStr(f.SocialType)},
		{"SortBy", ptrStr(f.SortBy)},
		{"SortOrder", ptrStr(f.SortOrder)},
		{"Page", ptrStr(f.Page)},
		{"Size", ptrStr(f.Size)},
	})
}

// Вспомогательные функции и типы
func structString(prefix string, fields []field) string {
	var parts []string
	for _, f := range fields {
		if f.value != "<nil>" && f.value != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", f.name, f.value))
		}
	}
	return prefix + "{" + strings.Join(parts, ", ") + "}"
}

func ifStr(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}
