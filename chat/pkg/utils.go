package pkg

import (
	"regexp"
)

/**
* Parses url with the given regular expression and returns the
* group values defined in the expression.
*
 */
func GetParams(regEx string, url string) (paramsMap map[string]string) {
	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func AddCurrentMessages(roomsMessages map[string][]string, room string, message string) {
	msgs := roomsMessages[room]
	if len(msgs) > 50 {
		msgs = msgs[1:]
	}
	msgs = append(msgs, message)

	mapMutex.RLock()
	roomsMessages[room] = msgs
	mapMutex.RUnlock()
}
