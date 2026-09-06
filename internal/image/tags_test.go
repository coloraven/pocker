package image_test

import (
	"reflect"
	"testing"

	"github.com/pocker/pocker/internal/image"
)

func TestNormalizeExtraNames(t *testing.T) {
	primary := "ghcr.io/username/image:tag"
	got, err := image.NormalizeExtraNames(primary, []string{
		"username/images:tag",
		"ghcr.io/username/image:tag", // duplicate of primary
		"username/images:tag",        // duplicate extra
		"myapp",                      // → myapp:latest
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"username/images:tag", "myapp:latest"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}
