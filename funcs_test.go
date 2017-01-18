package envsubst

import "testing"

func Test_len(t *testing.T) {
	got, want := toLen("Hello World"), "11"
	if got != want {
		t.Errorf("Expect len function to return %s, got %s", want, got)
	}
}

func Test_lower(t *testing.T) {
	got, want := toLower("Hello World"), "hello world"
	if got != want {
		t.Errorf("Expect lower function to return %s, got %s", want, got)
	}
}

func Test_lowerFirst(t *testing.T) {
	got, want := toLowerFirst("HELLO WORLD"), "hELLO WORLD"
	if got != want {
		t.Errorf("Expect lowerFirst function to return %s, got %s", want, got)
	}
	defer func() {
		if recover() != nil {
			t.Errorf("Expect empty string does not panic lowerFirst")
		}
	}()
	toLowerFirst("")
}

func Test_upper(t *testing.T) {
	got, want := toUpper("Hello World"), "HELLO WORLD"
	if got != want {
		t.Errorf("Expect upper function to return %s, got %s", want, got)
	}
}

func Test_upperFirst(t *testing.T) {
	got, want := toUpperFirst("hello world"), "Hello world"
	if got != want {
		t.Errorf("Expect upperFirst function to return %s, got %s", want, got)
	}
	defer func() {
		if recover() != nil {
			t.Errorf("Expect empty string does not panic upperFirst")
		}
	}()
	toUpperFirst("")
}

func Test_default(t *testing.T) {
	got, want := toDefault("Hello World", "Hola Mundo"), "Hello World"
	if got != want {
		t.Errorf("Expect default function uses variable value")
	}

	got, want = toDefault("", "Hola Mundo"), "Hola Mundo"
	if got != want {
		t.Errorf("Expect default function uses default value, when variable empty")
	}
}
