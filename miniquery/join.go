package miniquery

import "strings"

func Join(s []string) string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return s[0]
	}
	sb := strings.Builder{}
	for _, v := range s {
		if len(strings.TrimSpace(v)) == 0 {
			continue
		}
		if sb.Len() != 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString("(")
		sb.WriteString(v)
		sb.WriteString(")")
	}
	return sb.String()
}
