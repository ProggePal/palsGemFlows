package templating

import "testing"

func TestRenderString_ReplacesWithAndWithoutSpaces(t *testing.T) {
	memory := map[string]string{"get_topic": "cats"}

	got, err := RenderString("Topic: {{get_topic}}", memory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Topic: cats" {
		t.Fatalf("expected %q, got %q", "Topic: cats", got)
	}

	got, err = RenderString("Topic: {{ get_topic }}", memory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Topic: cats" {
		t.Fatalf("expected %q, got %q", "Topic: cats", got)
	}
}

func TestRenderString_MissingVariableErrors(t *testing.T) {
	_, err := RenderString("Hello {{ missing }}", map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
