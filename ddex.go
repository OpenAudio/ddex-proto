package ddex

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/OpenAudio/ddex-proto/gen/ddex/ern/v383"
	"github.com/OpenAudio/ddex-proto/gen/ddex/ern/v43"
	"github.com/OpenAudio/ddex-proto/gen/ddex/ern/v432"
	"github.com/OpenAudio/ddex-proto/gen/ddex/mead/v11"
	"github.com/OpenAudio/ddex-proto/gen/ddex/pie/v10"
)

// Versioned type aliases for discoverability of pure XML types
type (
	// ERN v4.3 - Main message types
	NewReleaseMessageV43   = ernv43.NewReleaseMessage
	PurgeReleaseMessageV43 = ernv43.PurgeReleaseMessage

	// ERN v3.8.3 - Main message types
	NewReleaseMessageV383   = ernv383.NewReleaseMessage
	PurgeReleaseMessageV383 = ernv383.PurgeReleaseMessage
	CatalogListMessageV383  = ernv383.CatalogListMessage

	// ERN v4.3.2 - Main message types
	NewReleaseMessageV432   = ernv432.NewReleaseMessage
	PurgeReleaseMessageV432 = ernv432.PurgeReleaseMessage

	// MEAD v1.1 types
	MeadMessageV11 = meadv11.MeadMessage

	// PIE v1.0 types
	PieMessageV10        = piev10.PieMessage
	PieRequestMessageV10 = piev10.PieRequestMessage
)

// ERNVersion represents a supported ERN version
type ERNVersion string

const (
	ERNv43  ERNVersion = "43"
	ERNv383 ERNVersion = "383"
	ERNv432 ERNVersion = "432"
)

// ERNMessage represents any ERN message type
type ERNMessage interface {
	// All ERN messages can be marshaled to XML
	xml.Marshaler
}

// DetectERNVersion detects the ERN version from XML content
func DetectERNVersion(xmlData []byte) (ERNVersion, error) {
	xmlStr := string(xmlData)

	// Look for ERN namespace patterns
	ernPattern := regexp.MustCompile(`xmlns(?::\w+)?="http://ddex\.net/xml/ern/v?(\d+(?:\.\d+)*)`)
	matches := ernPattern.FindStringSubmatch(xmlStr)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not detect ERN version from XML")
	}

	version := strings.ReplaceAll(matches[1], ".", "")

	switch version {
	case "43":
		return ERNv43, nil
	case "383":
		return ERNv383, nil
	case "432":
		return ERNv432, nil
	default:
		return "", fmt.Errorf("unsupported ERN version: %s", version)
	}
}

// ParseERN automatically detects version and parses ERN XML to appropriate message type
func ParseERN(xmlData []byte) (ERNMessage, ERNVersion, error) {
	version, err := DetectERNVersion(xmlData)
	if err != nil {
		return nil, "", err
	}

	message, err := ParseERNWithVersion(xmlData, version)
	return message, version, err
}

// ParseERNWithVersion parses ERN XML to specific version message type
func ParseERNWithVersion(xmlData []byte, version ERNVersion) (ERNMessage, error) {
	xmlStr := string(xmlData)

	// Determine message type (handle both namespaced and non-namespaced forms)
	if strings.Contains(xmlStr, "NewReleaseMessage") {
		return parseNewReleaseMessage(xmlData, version)
	} else if strings.Contains(xmlStr, "PurgeReleaseMessage") {
		return parsePurgeReleaseMessage(xmlData, version)
	}

	return nil, fmt.Errorf("unknown ERN message type")
}

func parseNewReleaseMessage(xmlData []byte, version ERNVersion) (ERNMessage, error) {
	switch version {
	case ERNv43:
		var msg NewReleaseMessageV43
		err := xml.Unmarshal(xmlData, &msg)
		return &msg, err
	case ERNv383:
		var msg NewReleaseMessageV383
		err := xml.Unmarshal(xmlData, &msg)
		return &msg, err
	case ERNv432:
		var msg NewReleaseMessageV432
		err := xml.Unmarshal(xmlData, &msg)
		return &msg, err
	default:
		return nil, fmt.Errorf("unsupported ERN version: %s", version)
	}
}

func parsePurgeReleaseMessage(xmlData []byte, version ERNVersion) (ERNMessage, error) {
	switch version {
	case ERNv43:
		var msg PurgeReleaseMessageV43
		err := xml.Unmarshal(xmlData, &msg)
		return &msg, err
	case ERNv383:
		var msg PurgeReleaseMessageV383
		err := xml.Unmarshal(xmlData, &msg)
		return &msg, err
	case ERNv432:
		var msg PurgeReleaseMessageV432
		err := xml.Unmarshal(xmlData, &msg)
		return &msg, err
	default:
		return nil, fmt.Errorf("unsupported ERN version: %s", version)
	}
}
