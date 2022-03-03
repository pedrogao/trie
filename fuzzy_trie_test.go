package trie

import (
	"testing"
)

func TestFuzzyTrie_Delete(t *testing.T) {
	type callbackFunc func() bool

	trie := NewFuzzyTrie()

	trie.Put("/usr", "pedro")
	trie.Put("/usr/age", "25")
	trie.Put("/usr/gender", "male")
	trie.Put("/usr/garden", "qh")
	trie.Put("/usr/colors", "bucket")
	trie.Put("/usr/colors/1", "black")
	trie.Put("/usr/colors/2", "red")
	trie.Put("/usr/colors/3", "white")

	tests := []struct {
		name     string
		arg      string
		want     bool
		callback callbackFunc
	}{
		{
			name: "delete /usr/age/aa",
			arg:  "/usr/age/aa",
			want: false,
			callback: func() bool {
				got := trie.Get("/usr/age")
				return got == "25"
			},
		},
		{
			name: "delete /usr/colors/*",
			arg:  "/usr/colors/*",
			want: true,
			callback: func() bool {
				got := trie.Get("/usr/colors/1")
				return got == nil
			},
		},
		{
			name: "delete /usr/g*",
			arg:  "/usr/g*",
			want: true,
			callback: func() bool {
				got := trie.Get("/usr/gender")
				return got == nil
			},
		},
		{
			name: "delete /usr",
			arg:  "/usr",
			want: true,
			callback: func() bool {
				got := trie.Get("/usr")
				return got == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trie.Delete(tt.arg); got != tt.want {
				t.Errorf("Delete() = %v, want %v", got, tt.want)
			}
			if tt.callback != nil && !tt.callback() {
				t.Errorf("Delete() = %v, callback err", tt.arg)
			}
		})
	}
}

func TestFuzzyTrie_Walk(t *testing.T) {
	trie := NewFuzzyTrie()

	trie.Put("/usr", "pedro")
	trie.Put("/usr/age", "25")
	trie.Put("/usr/gender", "male")
	trie.Put("/usr/garden", "qh")
	trie.Put("/usr/colors", "bucket")
	trie.Put("/usr/colors/1", "black")
	trie.Put("/usr/colors/2", "red")
	trie.Put("/usr/colors/3", "white")

	tests := []struct {
		name string
		want int
	}{
		{
			name: "walk",
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := 0
			if err := trie.Walk(func(key string, value interface{}) error {
				t.Logf("k: %s, v: %v\n", key, value)
				l += 1
				return nil
			}); err != nil {
				t.Errorf("Walk() err")
			}
			if l != tt.want {
				t.Errorf("Walk() err, got=%d, expected=%d", l, tt.want)
			}
		})
	}
}

func TestFuzzyTrie_WalkPath(t *testing.T) {
	trie := NewFuzzyTrie()

	trie.Put("/usr", "pedro")
	trie.Put("/usr/age", "25")
	trie.Put("/usr/gender", "male")
	trie.Put("/usr/garden", "qh")
	trie.Put("/usr/colors", "bucket")
	trie.Put("/usr/colors/1", "black")
	trie.Put("/usr/colors/2", "red")
	trie.Put("/usr/colors/3", "white")

	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "walk /usr/colors",
			arg:  "/usr/colors",
			want: 2,
		},
		{
			name: "walk /usr/colors/*",
			arg:  "/usr/colors/*",
			want: 5,
		},
		{
			name: "walk /usr/g*",
			arg:  "/usr/g*",
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := 0
			if err := trie.WalkPath(tt.arg, func(key string, value interface{}) error {
				t.Logf("k: %s, v: %v\n", key, value)
				l += 1
				return nil
			}); err != nil {
				t.Errorf("WalkPath() err")
			}
			if l != tt.want {
				t.Errorf("WalkPath() err, got=%d, expected=%d", l, tt.want)
			}
		})
	}
}

func TestFuzzyTrie_PutAndGet(t *testing.T) {
	trie := NewFuzzyTrie()

	got := trie.Get("/usr")
	assertNil(t, got)

	ok := trie.Put("/usr", "pedro")
	assertTrue(t, ok)

	got = trie.Get("/usr")
	assertNotNil(t, got)

	ok = trie.Put("/usr/age", "25")
	assertTrue(t, ok)

	ok = trie.Put("/usr/gender", "male")
	assertTrue(t, ok)

	ok = trie.Put("/usr/garden", "qh")
	assertTrue(t, ok)

	ok = trie.Put("/usr/colors", "bucket")
	assertTrue(t, ok)

	got = trie.Get("/usr/colors")
	assertNotNil(t, got)

	ok = trie.Put("/usr/colors/1", "black")
	assertTrue(t, ok)

	ok = trie.Put("/usr/colors/2", "red")
	assertTrue(t, ok)

	got = trie.Get("/usr/colors/3")
	assertNil(t, got)

	got = trie.Get("/usr/colors/1")
	assertTrue(t, got == "black")

	ok = trie.Put("/usr/colors/3", "white")
	assertTrue(t, ok)
}

func assertFalse(t *testing.T, b bool) {
	if b {
		t.Fail()
	}
}

func assertTrue(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func assertNil(t *testing.T, val interface{}) {
	if val != nil {
		t.Errorf("%v not nil\n", val)
		t.Fail()
	}
}

func assertNotNil(t *testing.T, val interface{}) {
	if val == nil {
		t.Errorf("%v nil\n", val)
		t.Fail()
	}
}
