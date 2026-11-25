// Package onefix has an import one rule's fix rewrites and a function another rule's fix does.
package onefix

import str "strings"

func upper(s string) string { return str.ToUpper(s) }

func lower(s string) string { return str.ToLower(s) }
