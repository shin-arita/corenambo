package registrationurl

import "testing"

func TestStaticBuilderBuild(t *testing.T) {
	b := StaticBuilder{
		FrontendBaseURL: "http://example.com",
	}

	url := b.Build("token")

	expected := "http://example.com/user-registration/verify?token=token"

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
