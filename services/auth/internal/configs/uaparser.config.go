package configs

import "github.com/ua-parser/uap-go/uaparser"

func InitializeUserAgentParser() *uaparser.Parser {
	parser := uaparser.NewFromSaved()
	return parser
}
