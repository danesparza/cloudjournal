package token

import "strings"

// Replace replaces each instance of a token with its token value
func Replace(source string, tokens map[string]string) string {

	retval := source

	//	Replace all found keys with their values:
	for key, value := range tokens {
		retval = strings.ReplaceAll(retval, key, value)
	}

	return retval
}
