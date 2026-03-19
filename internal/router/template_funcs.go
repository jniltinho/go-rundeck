package router

import (
	"fmt"
	"path"
	"strings"
	"text/template"
)

// templateFuncMap returns the common template functions used across the application.
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"base":  path.Base,
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"fmtDur": func(p *float64) string {
			if p == nil {
				return "—"
			}
			return fmt.Sprintf("%.1fs", *p)
		},
		"isAdmin": func(role interface{}) bool {
			return fmt.Sprintf("%v", role) == "admin"
		},
		"hasRole": func(userRole interface{}, allowedRoles ...string) bool {
			r := fmt.Sprintf("%v", userRole)
			for _, allowed := range allowedRoles {
				if r == allowed {
					return true
				}
			}
			return false
		},
		"eqString": func(v1, v2 interface{}) bool {
			return fmt.Sprintf("%v", v1) == fmt.Sprintf("%v", v2)
		},
		"keyTypeClass": func(keyType interface{}) string {
			if fmt.Sprintf("%v", keyType) == "private_key" {
				return "bg-[#FFD600] text-[#0C0C0C]"
			}
			return "bg-blue-100 text-blue-800"
		},
	}
}
