package m3u

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ResourceType int

const (
	ResourceTypeUnknown ResourceType = iota
	ResourceTypeFile
	ResourceTypeURL
)

type EnvelopType int

const (
	EnvelopTypeUnknown EnvelopType = iota
	EnvelopTypeEOF
	EnvelopTypeMetadata
)

type Parser struct {
	resource     string
	resourceType ResourceType
	content      io.ReadCloser
	regx         *regexp.Regexp
}

type Envelop struct {
	Type      EnvelopType
	URL       string
	OtherName string
	Group     string
	Name      string
	Logo      string
	ID        string
	RawValue  string
	Err       error
}

func NewParser(resource string, resourceType ResourceType) (*Parser, error) {
	regx, err := regexp.Compile(`(?P<prefix>\#EXTINF\:-1) (tvg-id="(?P<id>.*?)") (tvg-name="(?P<name>.*?)") (tvg-logo="(?P<logo>.*?)") (group-title="(?P<group>.*?)"),(?P<additional_name>.*)`)
	if err != nil {
		return nil, err
	}

	var content io.ReadCloser
	switch resourceType {
	case ResourceTypeFile:
		content, err = loadFile(resource)
	case ResourceTypeURL:
		content, err = loadURL(resource)
	default:
		return nil, fmt.Errorf("resource type not supported: %d", resourceType)
	}

	return &Parser{
		resource:     resource,
		resourceType: resourceType,
		content:      content,
		regx:         regx,
	}, err
}

func (p *Parser) Parse(ch chan Envelop) {
	go func(p *Parser) {
		defer p.content.Close()
		defer close(ch)

		var lines int
		var envelop Envelop
		scanner := bufio.NewScanner(p.content)
		for scanner.Scan() {
			value := scanner.Text()
			if lines == 0 && value != "#EXTM3U" {
				ch <- Envelop{
					RawValue: value,
					Err:      fmt.Errorf("invalid header found: %s", value),
				}
				break
			}

			match := p.regx.FindStringSubmatch(value)
			if len(match) == 0 {
				if len(value) > 0 && !strings.HasPrefix(value, "#EXT") {
					envelop.URL = value
					ch <- envelop
				}
				lines++
				continue
			}

			envelop = Envelop{}
			envelop.RawValue = value
			envelop.Type = EnvelopTypeMetadata
			for i, name := range p.regx.SubexpNames() {
				if len(name) == 0 {
					continue
				}

				switch name {
				case "id":
					envelop.ID = match[i]
				case "name":
					envelop.Name = match[i]
				case "logo":
					envelop.Logo = match[i]
				case "group":
					envelop.Group = match[i]
				case "additional_name":
					envelop.OtherName = match[i]
				}
			}

			lines++
		}

		ch <- Envelop{
			Type: EnvelopTypeEOF,
		}
	}(p)
}

func loadFile(filePath string) (io.ReadCloser, error) {
	return os.Open(filePath)
}

func loadURL(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("load URL '%s' failed due to HTTP status code: %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}
