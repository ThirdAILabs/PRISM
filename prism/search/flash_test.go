package search_test

import (
	"io"
	"os"
	"path/filepath"
	"prism/prism/search"
	"strings"
	"testing"
)

func init() {
	const licensePath = "../../.test_license/thirdai.license"
	if err := search.SetLicensePath(licensePath); err != nil {
		panic(err)
	}
}

func TestFlash(t *testing.T) {
	flash, err := search.NewFlash()
	if err != nil {
		t.Fatal(err)
	}

	defer flash.Free()

	data := []string{"entity", "this is a first sample", "completely different text here", "apples are a good fruit", "hopefully this is enough data for the test"}

	filename := filepath.Join(t.TempDir(), "test.csv")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(file, strings.NewReader(strings.Join(data, "\n"))); err != nil {
		t.Fatal(err)
	}

	if err := flash.Train(filename); err != nil {
		t.Fatal(err)
	}

	results, err := flash.Predict("this is enough", 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) < 1 || results[0] != data[4] {
		t.Fatalf("incorrect results: %v", results)
	}
}
