package bot

import (
	"regexp"
	"strings"
	"strconv"
	"github.com/eefret/hsapi"
	"bytes"
	"fmt"
)

func parseParams(message string, prefix string) (name string, params map[string]interface{}) {
	params = make(map[string]interface{})

	params[PARAM_COMMAND] = prefix
	message = strings.Replace(message, prefix, "", 1)
	if !strings.Contains(message, "[") {
		name = strings.TrimSpace(message)
		return
	}

	name = message[0:strings.Index(message, "[")]
	name = strings.TrimSpace(name)

	re := regexp.MustCompile(`\[(.*?)\]`)
	values := re.FindAllString(message, MAX_PARAMS)
	for _, value := range values {
		value = strings.Trim(value, "[]")

		if index, err := strconv.Atoi(value); err == nil {
			params[PARAM_INDEX] = index
			continue
		}
		//TODO: parse sound type params
	}
	return
}

func createMultiCardError(cards []hsapi.Card, cardname string, command string) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("I found a lot of cards with the name `%s`: \n", cardname))
	for k, v := range cards {
		buffer.WriteString(fmt.Sprintf("%d- %s [type:%s, cost:%d]\n", k, v, v.Type, v.Cost))
	}
	buffer.WriteString(fmt.Sprintf("Select which one is yours with `%s %s[position]` or \n", command, cardname))
	buffer.WriteString("redo the query with a more specific name... ")
	return buffer.String()
}
