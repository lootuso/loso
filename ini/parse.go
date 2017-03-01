package ini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Value map[string]interface{}

var (
	sectionRegex = regexp.MustCompile(`^\[(.*)\]$`)
	assignRegex  = regexp.MustCompile(`^([^=]+)=(.*)$`)
)

var (
	FOUND_ERROR = errors.New("not found")
	TYPE_ERROR  = errors.New("unknow type")
)

// SyntaxError is returned when there is a syntax error in an INI file.
type SyntaxError struct {
	Line   int
	Source string // The contents of the erroneous line, without leading or trailing whitespace
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

type Ini struct {
	storage map[string]Value
}

func (i *Ini) Section(session string) Value {
	if session, ok := i.storage[session]; ok {
		return session
	}
	return nil
}

func (i *Ini) Get(session, key string) (interface{}, error) {
	s := i.Section(session)
	if s != nil {
		if v, ok := s[key]; ok {
			return v, nil
		}
	}
	return nil, FOUND_ERROR
}

func (i *Ini) String(session, key string) (string, error) {
	if val, err := i.Get(session, key); err == nil {
		if v, ok := val.(string); ok {
			return v, nil
		} else {
			return "", TYPE_ERROR
		}
	}
	return "", FOUND_ERROR
}

func (i *Ini) Integer(session, key string) (int, error) {
	if value, err := i.Get(session, key); err == nil {
		var v int
		var err error
		switch val := value.(type) {
		case string:
			v, err = strconv.Atoi(val)
			if err != nil {
				v = 0
			}
		case int:
			v = val
		default:
			v = 0
			err = TYPE_ERROR
		}
		if err == nil {
			return v, nil
		}
		return 0, err
	}
	return 0, FOUND_ERROR
}

func (i *Ini) parse(reader io.Reader) error {
	section := ""
	lineNum := 0
	var err error
	bufferReader := bufio.NewReader(reader)
	for done := false; !done; {
		var line string
		if line, err = bufferReader.ReadString('\n'); err != nil {
			if err == io.EOF {
				done = true
			} else {
				return err
			}
		}
		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}
		if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := strings.TrimSpace(groups[1]), strings.TrimSpace(groups[2])
			i.storage[section][key] = val
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			section = strings.TrimSpace(groups[1])
			i.storage[section] = make(Value)
		} else {
			return SyntaxError{lineNum, line}
		}
	}
	return nil
}

func NewIni(source string) (*Ini, error) {
	ini := &Ini{
		storage: make(map[string]Value),
	}
	var err error
	if _, err = os.Stat(source); err == nil {
		if reader, err := os.Open(source); err == nil {
			defer reader.Close()
			if err := ini.parse(reader); err == nil {
				return ini, nil
			}
		}
		return nil, err
	} else {
		reader := strings.NewReader(source)
		if err := ini.parse(reader); err == nil {
			return ini, nil
		} else {
			return nil, err
		}
	}
}
