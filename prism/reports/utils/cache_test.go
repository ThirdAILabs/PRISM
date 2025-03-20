package utils_test

import (
	"path/filepath"
	"prism/prism/reports/utils"
	"testing"
)

func TestCache(t *testing.T) {
	type cachedData struct {
		Name string
		Cnt  int
	}

	cache, err := utils.NewCache[cachedData]("somebucket", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	if cache.Lookup("xyz") != nil {
		t.Fatal("should be no cached result")
	}

	cache.Update("xyz", cachedData{"xyz", 2})

	res1 := cache.Lookup("xyz")
	if res1 == nil || res1.Name != "xyz" || res1.Cnt != 2 {
		t.Fatal("invalid cached result")
	}

	if cache.Lookup("abc") != nil {
		t.Fatal("should be no cached result")
	}

	cache.Update("xyz", cachedData{"xyz-2", 5})

	res2 := cache.Lookup("xyz")
	if res2 == nil || res2.Name != "xyz-2" || res2.Cnt != 5 {
		t.Fatal("invalid cached result")
	}
}
