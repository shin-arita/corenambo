package registrationurl

import (
	"strings"
	"testing"
)

func TestStaticBuilderBuild(t *testing.T) {
	b := StaticBuilder{FrontendBaseURL: "http://example.com"}

	url := b.Build("token")

	expected := "http://example.com/registration/verify?token=token"
	if url != expected {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestNewStaticBuilder(t *testing.T) {
	b := NewStaticBuilder("http://example.com")

	if b == nil {
		t.Fatal("builder is nil")
	}
}

func TestStaticBuilderBuild_TokenAppearsInURL(t *testing.T) {
	b := StaticBuilder{FrontendBaseURL: "http://localhost:5173"}

	rawToken := "abc123XYZ"
	url := b.Build(rawToken)

	if !strings.Contains(url, "token="+rawToken) {
		t.Fatalf("raw token not found in URL: %s", url)
	}
}

func TestStaticBuilderBuild_URLNotEndingWithTokenEquals(t *testing.T) {
	b := StaticBuilder{FrontendBaseURL: "http://localhost:5173"}

	url := b.Build("sometoken")

	if strings.HasSuffix(url, "token=") {
		t.Fatalf("URL ends with empty token (token=): %s", url)
	}
}

func TestStaticBuilderBuild_EmptyTokenProducesEmptySuffix(t *testing.T) {
	b := StaticBuilder{FrontendBaseURL: "http://localhost:5173"}

	url := b.Build("")

	if !strings.HasSuffix(url, "token=") {
		t.Fatalf("expected URL to end with token= when empty, got: %s", url)
	}
}
