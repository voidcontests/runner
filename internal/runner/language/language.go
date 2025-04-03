package language

import (
	"errors"
)

var (
	ErrUnknownLanguage     = errors.New("runner: unknown language")
	ErrNotCompiledLanguage = errors.New("runner: language is not compiled")
)

const (
	C      = "c"
	Python = "python"
)

type Kind = uint8

const (
	Compiled Kind = 1 << iota
	Interpreted
)

var languageKinds = map[string]uint8{
	C:      Compiled,
	Python: Interpreted,
}

var languageExtensions = map[string]string{
	C:      "c",
	Python: "py",
}

type Language struct {
	Name      string
	Kind      Kind
	Extension string
}

func Get(name string) (Language, bool) {
	var l Language
	var ok bool

	l.Kind, ok = languageKinds[name]
	if !ok {
		return Language{}, false
	}

	l.Extension, ok = languageExtensions[name]
	if !ok {
		return Language{}, false
	}

	l.Name = name

	return l, true
}
