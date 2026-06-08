package monitorlog

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxLogValueLength = 512

func Event(component string, event string, fields ...any) {
	component = strings.TrimSpace(component)
	event = strings.TrimSpace(event)
	if component == "" || event == "" {
		return
	}

	var builder strings.Builder
	builder.WriteString(component)
	builder.WriteString(" monitor event=")
	builder.WriteString(formatValue(event))
	for index := 0; index+1 < len(fields); index += 2 {
		key := formatKey(fields[index])
		value := formatValue(fields[index+1])
		if key == "" || value == "" {
			continue
		}
		builder.WriteByte(' ')
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(value)
	}
	log.Print(builder.String())
}

func formatKey(value any) string {
	key := strings.TrimSpace(fmt.Sprint(value))
	if key == "" {
		return ""
	}
	var builder strings.Builder
	for _, character := range key {
		switch {
		case character >= 'a' && character <= 'z':
			builder.WriteRune(character)
		case character >= 'A' && character <= 'Z':
			builder.WriteRune(character)
		case character >= '0' && character <= '9':
			builder.WriteRune(character)
		case character == '_' || character == '-':
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func formatValue(value any) string {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return ""
		}
		return typed.UTC().Format(time.RFC3339Nano)
	case *time.Time:
		if typed == nil || typed.IsZero() {
			return ""
		}
		return typed.UTC().Format(time.RFC3339Nano)
	}

	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return ""
	}
	text = truncateLogValue(text)
	if isBareLogValue(text) {
		return text
	}
	return strconv.Quote(text)
}

func truncateLogValue(value string) string {
	if utf8.RuneCountInString(value) <= maxLogValueLength {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxLogValueLength]) + "..."
}

func isBareLogValue(value string) bool {
	for _, character := range value {
		if unicode.IsSpace(character) || character == '"' || character == '=' {
			return false
		}
	}
	return true
}
