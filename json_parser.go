// TODO
// change for loop into one taking byte sectioning of the string under account
// https://blog.golang.org/strings
// change json_object from a list of rows to a map of string ----- anyjs

package main

// lexer will take in a stringified json and return a slice of tokens

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type kind uint

const (
	str kind = iota
	bl
	fl
	in
	nl
	sx
	nn //signifies nil
)

type anyjs struct {
	kind  kind
	value interface{}
}

type json_row struct {
	name string
	anyjs
}

type json_object struct {
	values []json_row
}

func lex_string(s string) (anyjs, string) {
	json_string := ""
	if s[0] == '"' {
		s = s[1:]
	} else {
		return anyjs{nn, nil}, s
	}
	for _, c := range s {
		if c == '"' {
			return anyjs{str, json_string}, s[len(json_string)+1:]
		} else {
			json_string += string(c)
		}
	}
	return anyjs{nn, nil}, s
}

func lex_number(s string) (anyjs, string) {
	json_number := ""
	var number_characters = map[rune]bool{
		'-': true,
		'.': true,
		'e': true,
		'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true,
	}
	for _, c := range s {
		if _, ok := number_characters[c]; ok {
			json_number += string(c)
		} else {
			break
		}
	}
	rest := s[len(json_number):]
	if len(json_number) == 0 {
		return anyjs{nn, nil}, s
	}
	if strings.Contains(json_number, string('.')) {
		val, err := strconv.ParseFloat(json_number, 64)
		if err != nil {
			return anyjs{nn, nil}, s
		}
		return anyjs{fl, val}, rest
	}
	val, err := strconv.ParseInt(json_number, 10, 64)
	if err != nil {
		return anyjs{nn, nil}, s
	}
	return anyjs{in, val}, rest
}

func lex_bool(s string) (anyjs, string) {
	string_len := len(s)

	// 4 is length of TRUE
	if string_len >= 4 && s[:4] == "true" {
		return anyjs{bl, true}, s[4:]
	} else if string_len >= 5 && s[:5] == "false" {
		// 5 is length of FALSE
		return anyjs{bl, false}, s[5:]
	}
	return anyjs{nn, nil}, s
}

func lex_null(s string) (anyjs, string) {
	string_len := len(s)

	if string_len >= 4 && s[:4] == "null" {
		return anyjs{nl, nil}, s[4:]
	}
	return anyjs{nn, nil}, s
}

func lex(s string) []anyjs {
	var JSON_WHITESPACE = map[byte]bool{
		'\t': true,
		'\n': true,
		' ':  true,
	}
	var JSON_SYNTAX = map[byte]bool{
		'{': true,
		'}': true,
		'[': true,
		']': true,
		',': true,
		':': true,
	}
	var tokens []anyjs
	for len(s) > 0 {
		var json_string anyjs
		var json_number anyjs
		var json_bool anyjs
		var json_null anyjs
		json_string, s = lex_string(s)
		if json_string.value != nil {
			tokens = append(tokens, json_string)
		}
		json_number, s = lex_number(s)
		if json_number.value != nil {
			tokens = append(tokens, json_number)
		}
		json_bool, s = lex_bool(s)
		if json_bool.value != nil {
			tokens = append(tokens, json_bool)
		}
		json_null, s = lex_null(s)
		if json_null.value != nil {
			tokens = append(tokens, json_null)
		}
		if _, ok := JSON_WHITESPACE[s[0]]; ok {
			s = s[1:]
		} else if _, ok := JSON_SYNTAX[s[0]]; ok {
			tokens = append(tokens, anyjs{sx, s[0]})
			s = s[1:]
		} else {
			fmt.Printf("Some unexpected token has occured")
			break
		}
	}
	return tokens
}

func parse_simple_array(tokens []anyjs) ([]json_object, error) {
	//assumes that it starts from without a bracket
	var ret []json_object
	for len(tokens) > 0 {
		val, size, err := parse_object(tokens)
		if err != nil {
			return nil, err
		}
		ret = append(ret, val)
		if tokens[size].kind != sx || (tokens[size].value.(uint8) != ',' && tokens[size].value.(uint8) != ']') {
			return nil, errors.New("something wrong with string format")
		}
		tokens = tokens[size+1:]
	}
	return ret, nil
}

func parse_object(tokens []anyjs) (json_object, int, error) {
	//assumes that starts from a curly brace
	var ret json_object
	if tokens[0].kind != sx || tokens[0].value.(uint8) != '{' {
		return json_object{nil}, 0, errors.New("Object doesn't start from a brace")
	}
	idx := 1
	for idx < len(tokens) {
		// on each iteration grab one row
		// then check if after row we get a comma
		// or a closing brace
		if (tokens[idx].kind != str && tokens[idx+1].kind != sx ) || tokens[idx+1].value.(uint8) != ':' {
			return json_object{nil}, 0, errors.New("Incorrectly formatted object row")
		}
		ret.values = append(ret.values, json_row{tokens[idx].value.(string), tokens[idx+2]})
		if tokens[idx+3].kind != sx {
			return json_object{nil}, 0, errors.New("Lacking a comma or closing brace after json row")
		} else if tokens[idx+3].value.(uint8) == ',' {
			idx = idx + 4
			continue
		} else if tokens[idx+3].value.(uint8) == '}' {
			idx = idx + 4
			return ret, idx, nil
		} else {
			return json_object{nil}, 0, errors.New("Some weird symbol after json row")
		}
	}
	return json_object{nil}, 0, errors.New("No closing object brace")
}

func parse(tokens []anyjs) ([]json_object, error) {
	t := tokens[0]
	if t.kind == sx && t.value.(uint8) == '[' {
		return parse_simple_array(tokens[1:])
	} else if t.kind == sx && t.value.(uint8) == '{' {
		ret, _, err := parse_object(tokens)
		return []json_object{ret}, err
	} else {
		fmt.Println("Badly formatted JSON string")
		return nil, errors.New("Badly formatted JSON string")
	}
}

func main() {
	f, err := os.Open("example.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			panic(err)
		}
	}()
	s, err := ioutil.ReadAll(f)
	lexed := lex(string(s))
	parsed, err := parse(lexed)
	if err != nil {
			fmt.Println(err)
	}
	for idx, elem := range parsed {
		fmt.Printf("\n\nElement number %d \n", idx)
		for _, row := range elem.values {
			switch row.kind {
			case str:
				fmt.Printf("%s : %s\n", row.name, row.value)
				break
			case fl:
				fmt.Printf("%s : %f\n", row.name, row.value)
				break
			case in:
				fmt.Printf("%s : %d\n", row.name, row.value)
				break
			case bl:
				fmt.Printf("%s : %t\n", row.name, row.value)
				break
			case nl:
				fmt.Printf("%s :null\n", row.name)
				break
			case sx:
				fmt.Printf("%s : %c\n", row.name, row.value)
				break
			}
		}
	}
}
