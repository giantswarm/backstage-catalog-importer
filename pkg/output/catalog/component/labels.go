package component

import "regexp"

// labelValue is what Backstage accepts as a label value: up to 63 characters
// of letters, digits, `-`, `_` and `.`, beginning and ending with a letter or
// digit. Versions such as `10.3.0` qualify; refs such as `dev:abc123` do not.
var labelValue = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9_.-]{0,61}[A-Za-z0-9])?$`)

// IsValidLabelValue reports whether v may be used as a Backstage label value.
// A value that fails must go into an annotation instead, or be dropped: an
// invalid label makes the catalog reject the whole entity.
func IsValidLabelValue(v string) bool {
	return labelValue.MatchString(v)
}
