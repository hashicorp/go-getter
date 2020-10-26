package getter

import (
	"regexp"
	"strings"
)

// FileMatcher allows to filter files based on their source and name
type FileMatcher = func(source, fullFilename string) bool

// WithFileMatcher adds a FileMatcher to the client
// Any kind of function that returns a boolean from file properties can be used
func WithFileMatcher(fm FileMatcher) func(*Client) error {
	return func(c *Client) error {
		c.FileMatches = fm
		return nil
	}
}

// WithFileMatcher adds a regex FileMatcher to the client
func WithRegexFileMatcher(regex string) func(*Client) error {
	return func(c *Client) error {
		expression, err := regexp.Compile(regex)
		c.FileMatches = func(source, fullFilename string) bool {
			return expression.MatchString(fullFilename) || expression.MatchString((strings.TrimPrefix(fullFilename, source)))
		}
		return err
	}
}
