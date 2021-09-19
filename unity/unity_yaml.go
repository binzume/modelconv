package unity

import (
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type YAMLDoc struct {
	Tag   string // tag:unity3d.com,2011:1
	refID string // unity fileID
	Body  []byte
}

func (d *YAMLDoc) Decode(dst interface{}) error {
	return yaml.Unmarshal(d.Body, dst)
}

type yamlSplitter struct {
	data []byte
	pos  int
	tags map[string]string
}

func ParseYamlDocuments(data []byte) []*YAMLDoc {
	s := yamlSplitter{data: data, tags: map[string]string{}}
	docStart := -0
	var doc *YAMLDoc
	var docs []*YAMLDoc

	for s.pos < len(data)-3 {
		if data[s.pos] == '%' && data[s.pos+1] == 'T' && data[s.pos+2] == 'A' && data[s.pos+3] == 'G' {
			s.pos += 4
			name := strings.Trim(s.readToken(), "!")
			value := s.readToken()
			s.tags[name] = value
		} else if data[s.pos] == '-' && data[s.pos+1] == '-' && data[s.pos+2] == '-' {
			s.pos += 3
			if doc != nil {
				doc.Body = data[docStart:s.pos]
				docs = append(docs, doc)
			}
			doc = &YAMLDoc{
				Tag:   s.getTag(),
				refID: strings.TrimPrefix(s.readToken(), "&"),
			}
			s.nextLine()
			docStart = s.pos
			continue
		}
		s.nextLine()
	}
	if doc != nil {
		doc.Body = data[docStart:s.pos]
		docs = append(docs, doc)
	}
	return docs
}

func (s *yamlSplitter) getTag() string {
	tag := s.readToken()
	if len(tag) > 0 && tag[0] == '!' {
		t := strings.SplitN(tag[1:], "!", 2)
		if v, ok := s.tags[t[0]]; ok {
			tag = v + tag[len(t[0])+2:]
		}
	}
	return tag
}

func (s *yamlSplitter) readToken() string {
	for s.pos < len(s.data) && s.data[s.pos] == ' ' {
		s.pos++
	}
	st := s.pos
	for s.pos < len(s.data) && s.data[s.pos] != '\n' && s.data[s.pos] != ' ' {
		s.pos++
	}
	return string(s.data[st:s.pos])
}

func (s *yamlSplitter) nextLine() int {
	for s.pos < len(s.data) {
		if s.data[s.pos] == '\n' {
			s.pos++
			break
		}
		s.pos++
	}
	return s.pos
}
