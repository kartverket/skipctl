package cmd

import (
	"testing"
)

func TestIsValidCommitRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		// Special ref
		{name: "HEAD", in: "HEAD", want: true},

		// Short SHAs (7..40)
		{name: "SHA_7_lower", in: "1a2b3c4", want: true},
		{name: "SHA_7_upper", in: "1A2B3C4", want: true},
		{name: "SHA_6_too_short", in: "1a2b3c", want: false},
		{name: "SHA_8", in: "1a2b3c4d", want: true},
		{name: "SHA_40", in: "0123456789abcdef0123456789abcdef01234567", want: true},
		{name: "SHA_41_too_long", in: "0123456789abcdef0123456789abcdef01234567a", want: false},
		{name: "SHA_mixed_case", in: "AbCdEf0123", want: true},

		// Branch/tag-like names allowed by [A-Za-z0-9._/-]+
		{name: "simple_branch", in: "main", want: true},
		{name: "feature_branch", in: "feature/cool-thing", want: true},
		{name: "dotted_branch", in: "release/1.2.3", want: true},
		{name: "underscores_allowed", in: "feature_with_under", want: true},
		{name: "nested_paths", in: "foo/bar/baz", want: true},
		{name: "hyphen_and_dot", in: "hotfix-1.2.3", want: true},
		{name: "starts_with_digit", in: "1-try", want: true},
		{name: "contains_slash_dash_dot_underscore", in: "a_b/c-d.e", want: true},

		// Edge cases for branch-like names
		{name: "single_dash", in: "-", want: true}, // allowed by regex, though git may reject in reality
		{name: "single_dot", in: ".", want: true},  // allowed by regex
		{name: "single_underscore", in: "_", want: true},
		{name: "single_slash", in: "/", want: true}, // allowed by regex
		{name: "leading_slash", in: "/start", want: true},
		{name: "trailing_slash", in: "end/", want: true},
		{name: "double_slash", in: "a//b", want: true},
		{name: "multiple_consecutive_dots", in: "a..b", want: true},
		{name: "ends_with_dot", in: "a.", want: true},

		// Invalid due to characters not in the class
		{name: "space_in_name", in: "feature new", want: false},
		{name: "tab_in_name", in: "feature\tnew", want: false},
		{name: "plus_sign", in: "feature+new", want: false},
		{name: "at_sign", in: "release@next", want: false},
		{name: "caret", in: "topic^", want: false},
		{name: "tilde", in: "topic~1", want: false},
		{name: "colon", in: "ns:branch", want: false},
		{name: "question_mark", in: "what?", want: false},
		{name: "asterisk", in: "fe*ature", want: false},
		{name: "open_bracket", in: "rel[1]", want: false},
		{name: "backslash", in: "foo\\bar", want: false},

		// Empty string
		{name: "empty", in: "", want: false},

		// Looks like a ref but too-long hex (not matching 7..40)
		{name: "hex_50", in: "0123456789abcdef0123456789abcdef0123456789abcdef", want: false},

		// HEAD variants
		{name: "head_exact", in: "HEAD", want: true},
		{name: "head_lowercase", in: "head", want: false},
		{name: "head_with_suffix_tilde_1", in: "HEAD~1", want: true},
		{name: "head_with_suffix_caret", in: "HEAD^", want: true},
		{name: "head_with_suffix_caret_2", in: "HEAD^2", want: true},

		// If using Option B
		{name: "head_reflog_index", in: "HEAD@{1}", want: true},
		{name: "head_reflog_invalid_empty", in: "HEAD@{}", want: false},
		{name: "head_reflog_negative", in: "HEAD@{-1}", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsValidCommitRef(tc.in)
			if got != tc.want {
				t.Fatalf("IsValidCommitRef(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
