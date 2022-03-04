package goimports

import (
	"fmt"
	"log"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=SectionType
type SectionType int

const (
	_ SectionType = iota
	Standard
	Default
	Prefix
	NewLine
	Comment
)

type Sections []Section

type Section struct {
	Type  SectionType
	Value string
}

func (s Section) Match(v string) bool {
	switch s.Type {
	case Standard:
		return isStandardPackage(v)
	case Default:
		return false
	case Prefix:
		return strings.HasPrefix(v, s.Value)
	}
	return false
}

func (s Section) IsDefault() bool {
	return s.Type == Default
}

func (s Section) String() string {
	return fmt.Sprintf("%s(%s)", s.Type, s.Value)
}

func (s Sections) String() string {
	var builder strings.Builder
	for _, section := range s {
		builder.WriteString(section.String())
	}
	return builder.String()
}

func ParseSection(v string) Section {
	switch {
	case v == Standard.String():
		return Section{Type: Standard}
	case v == Default.String():
		return Section{Type: Default}
	case v == NewLine.String():
		return Section{Type: NewLine}
	case strings.HasPrefix(v, Prefix.String()):
		return Section{Type: Prefix, Value: extractValue(v)}
	case strings.HasPrefix(v, Standard.String()):
		return Section{Type: Prefix, Value: extractValue(v)}
	}
	log.Println("unknown section type, ignored")
	return Section{}
}

func extractValue(v string) string {
	start := strings.Index(v, "(")
	end := strings.LastIndex(v, ")")
	return v[start+1:end]
}

func (s *Sections) Set(v string) error {
	*s = append(*s, ParseSection(v))
	return nil
}

func (s Sections) DefaultIndex() int {
	for i, section := range s {
		if section.IsDefault() {
			return i
		}
	}
	return -1
}

var (
	sections Sections
)

func init() {
	Analyzer.Flags.Var(&sections, "section", "section")
}
