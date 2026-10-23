package locales

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/i18n"
)

func TestCatalogCompleteness(t *testing.T) {
	want := []string{"en", "ja", "ko", "zh-Hans", "zh-Hant"}
	if !slices.Equal(Catalog.Languages(), want) {
		t.Fatalf("languages = %v", Catalog.Languages())
	}
	read := func(tag string) map[string]json.RawMessage {
		t.Helper()
		data, err := files.ReadFile(tag + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var messages map[string]json.RawMessage
		if err := json.Unmarshal(data, &messages); err != nil {
			t.Fatal(err)
		}
		return messages
	}
	source := read("en")
	for _, tag := range want {
		messages := read(tag)
		for key := range source {
			if strings.HasPrefix(key, "@") {
				continue
			}
			var text string
			if err := json.Unmarshal(messages[key], &text); err != nil || text == "" {
				t.Errorf("%s is missing %s", tag, key)
			}
		}
	}
}

func TestLanguageMatchingAndPlaceholders(t *testing.T) {
	for _, tc := range []struct{ preferred, matched, settings string }{
		{"en-GB", "en", "Settings"},
		{"zh-CN", "zh-Hans", "设置"},
		{"zh-TW", "zh-Hant", "設定"},
		{"zh-HK", "zh-Hant", "設定"},
		{"ja-JP", "ja", "設定"},
		{"ko-KR", "ko", "설정"},
	} {
		if got := i18n.Match([]string{tc.preferred}, Catalog.Languages()); got != tc.matched {
			t.Errorf("%s matched %s", tc.preferred, got)
		}
		if got := In(tc.preferred).Settings(); got != tc.settings {
			t.Errorf("%s settings = %s", tc.preferred, got)
		}
	}
	if got := In("fr").Settings(); got != "Settings" {
		t.Fatalf("unsupported language fallback = %s", got)
	}
}
