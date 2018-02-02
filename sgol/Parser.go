package sgol

import (
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	Clauses  []string `json:"clauses" hcl:"clauses"`
	Entities []string `json:"entities" hcl:"entities"`
	Edges    []string `json:"edges" hcl:"edges"`
}

func (p *Parser) Rejoin(block []string) string {
	text := ""

	for _, token := range block {
		if strings.Contains(token, " ") && !(strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"")) {
			text += "\"" + token + "\""
		} else {
			text += token + " "
		}
	}

	return text
}

func (p *Parser) ParseEntities(text string) []string {
	if len(text) > 0 {
		if strings.HasPrefix(text, "$") {
			if len(text) > 1 {
				if strings.Contains(text, ",") {
					parts := strings.Split(text, ",")
					result := make([]string, len(parts))
					for i, x := range parts {
						result[i] = x[1:len(x)]
					}
					return result
				} else {
					return []string{text[1:len(text)]}
				}
			} else {
				return p.Entities
			}
		} else if text == "_" || "text" == "-" {
			return []string{}
		} else {
			return []string{text}
		}
	} else {
		return []string{}
	}
}

func (p *Parser) ParseQueryFunctions(text string) ([]QueryFunction, error) {

	functions := make([]QueryFunction, 0)
	if len(text) > 0 {

		re, err := regexp.Compile("(\\s*)(?P<name>([a-zA-Z]+))(\\s*)\\((\\s*)(?P<args>(.)*?)(\\s*)\\)(\\s*)")
		if err != nil {
			return functions, err
		}

		matches := re.FindAllStringSubmatch(strings.TrimSpace(text), -1)
		for _, m := range matches {
			g := map[string]string{}
			for i, name := range re.SubexpNames() {
				if i != 0 {
					g[name] = m[i]
				}
			}
			fn := QueryFunction{Name: strings.TrimSpace(g["name"])}
			if args_text, ok := g["args"]; ok {
				if len(args_text) > 0 {
					args := []string{}

					re2, err := regexp.Compile("(\\s*)(?P<value>((\"([^\"]+?)\")|([^,\\s]+)))(\\s*)")
					if err != nil {
						return functions, err
					}

					matches2 := re.FindAllStringSubmatch(args_text, -1)
					for _, m2 := range matches2 {
						g2 := map[string]string{}
						for i, name := range re2.SubexpNames() {
							if i != 0 {
								g2[name] = m2[i]
							}
						}
						if value, ok := g2["value"]; ok {
							if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
								args = append(args, value[1:len(value)-1])
							} else {
								args = append(args, value)
							}
						}
					}
					fn.Args = args
				}
			}
			functions = append(functions, fn)
		}

	}
	return functions, nil
}

func (p *Parser) ParseFilterFunctions(text string) ([]FilterFunction, error) {
	filterFunctions := []FilterFunction{}
	queryFunctions, err := p.ParseQueryFunctions(text)
	if err != nil {
		return filterFunctions, err
	}
	if len(queryFunctions) > 0 {
		for _, qf := range queryFunctions {
			ff := FilterFunction{
				Selection: []string{qf.Args[0]},
				// TBD Add Predicates
			}
			filterFunctions = append(filterFunctions, ff)
		}
		return filterFunctions, nil
	} else {
		return []FilterFunction{}, nil
	}
}

func (p *Parser) ParseTokens(text string) ([]string, error) {
	tokens := []string{}

	re, err := regexp.Compile("(?P<token>((\"([^\"]+)\")|(\\S+)))")
	if err != nil {
		return tokens, err
	}

	matches := re.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		g := map[string]string{}
		for i, name := range re.SubexpNames() {
			if i != 0 {
				g[name] = m[i]
			}
		}
		tokens = append(tokens, strings.TrimSpace(g["token"]))
	}

	return tokens, nil

}

func (p *Parser) ParseBlocks(tokens []string) [][]string {
	blocks := [][]string{}
	block := []string{}

	for _, token := range tokens {
		token_uc := strings.ToUpper(token)
		if StringSliceContains(p.Clauses, token_uc) {
			if len(block) > 0 {
				blocks = append(blocks, block)
			}
			block = []string{token_uc}
		} else {
			block = append(block, token)
		}
	}

	if len(block) > 0 {
		blocks = append(blocks, block)
	}

	return blocks
}

func (p *Parser) ParseSelect(block []string) (OperationSelect, error) {
	op := OperationSelect{}

	if block[1] == "$" {
		op.Edges = p.Edges
	}

	if strings.HasPrefix(block[1], "$") {
		op.Entities = p.ParseEntities(block[1])
	} else {
		op.Entities = p.Entities
		op.Seeds = strings.Split(block[1], ",")
	}

	if len(block) > 2 {
		if strings.ToUpper(block[2]) == "UPDATE" {
			op.UpdateKey = block[3]
			if len(block) > 6 {
				filterFunctions, err := p.ParseFilterFunctions(p.Rejoin(block[7:len(block)]))
				if err != nil {
					return op, err
				}
				if len(filterFunctions) > 0 {
					op.UpdateFilterFunctions = map[string][]FilterFunction{}
					op.UpdateFilterFunctions[block[5][1:len(block[5])]] = filterFunctions
				}
			}
		} else if strings.ToUpper(block[2]) == "FILTER" {
			if StringSliceContains(block[3:len(block)], "UPDATE") {
				blockUpdateIndex := StringSliceIndex(block, "UPDATE")
				filterFunctions, err := p.ParseFilterFunctions(p.Rejoin(block[5:blockUpdateIndex]))
				if err != nil {
					return op, err
				}
				if len(filterFunctions) > 0 {
					op.UpdateFilterFunctions = map[string][]FilterFunction{}
					op.UpdateFilterFunctions[block[3][1:len(block[3])]] = filterFunctions
				}
				op.UpdateKey = block[blockUpdateIndex+1]
				if len(block) > blockUpdateIndex+2 {
					blockUpdateGroupIndex := blockUpdateIndex + 3
					updateFilterFunctions, err := p.ParseFilterFunctions(p.Rejoin(block[blockUpdateGroupIndex+2 : len(block)]))
					if err != nil {
						return op, err
					}
					if len(filterFunctions) > 0 {
						op.UpdateFilterFunctions = map[string][]FilterFunction{}
						op.UpdateFilterFunctions[block[blockUpdateGroupIndex][1:len(block[blockUpdateGroupIndex])]] = updateFilterFunctions
					}
				}
			} else {
				if len(block) > 5 && block[4] == "WITH" {
					filterFunctions, err := p.ParseFilterFunctions(p.Rejoin(block[5:len(block)]))
					if err != nil {
						return op, err
					}
					if len(filterFunctions) > 0 {
						op.UpdateFilterFunctions = map[string][]FilterFunction{}
						op.UpdateFilterFunctions[block[3][1:len(block[3])]] = filterFunctions
					}
				} else {
					filterFunctions, err := p.ParseFilterFunctions(p.Rejoin(block[3:len(block)]))
					if err != nil {
						return op, err
					}
					if len(filterFunctions) > 0 {
						op.UpdateFilterFunctions = map[string][]FilterFunction{}
						for _, entity := range p.Entities {
							op.UpdateFilterFunctions[entity] = filterFunctions
						}
					}
				}
			}
		}
	}

	return op, nil
}

func (p *Parser) ParseNav(block []string) (OperationNav, error) {
	op := OperationNav{}
	return op, nil
}

func (p *Parser) ParseHas(block []string) (OperationHas, error) {
	op := OperationHas{}
	return op, nil
}

func (p *Parser) ParseOperations(blocks [][]string) ([]Operation, string, error) {
	operations := make([]Operation, 0)
	output_type := ""
	for _, block := range blocks {
		switch block[0] {
		case "INIT":
			operations = append(operations, OperationInit{Key: block[1]})
		case "ADD":
			operations = append(operations, OperationAdd{Key: block[1]})
		case "DISCARD":
			operations = append(operations, OperationDiscard{Key: block[1]})
		case "RELATE":
			operations = append(operations, OperationRelate{Keys: []string{block[1]}})
		case "LIMIT":
			limit_int, err := strconv.Atoi(block[1])
			if err != nil {
				return operations, output_type, err
			}
			operations = append(operations, OperationLimit{Limit: limit_int})
		case "SELECT":
			op, err := p.ParseSelect(block)
			if err != nil {
				return operations, output_type, err
			}
			operations = append(operations, op)
		case "NAV":
			op, err := p.ParseNav(block)
			if err != nil {
				return operations, output_type, err
			}
			operations = append(operations, op)
		case "HAS":
			op, err := p.ParseHas(block)
			if err != nil {
				return operations, output_type, err
			}
			operations = append(operations, op)
		case "FETCH":
			operations = append(operations, OperationFetch{Key: block[1]})
		case "SEED":
			operations = append(operations, OperationSeed{})
		case "RUN":
			operations = append(operations, OperationRun{})
		case "OUTPUT":
			output_type = block[1]
		}
	}
	return operations, output_type, nil
}

func (p *Parser) ParseQuery(q string) ([]Operation, string, error) {
	tokens, err := p.ParseTokens(q)
	if err != nil {
		return []Operation{}, "", err
	}

	blocks := p.ParseBlocks(tokens)

	operations, outputType, err := p.ParseOperations(blocks)
	if err != nil {
		return operations, outputType, err
	}

	return operations, outputType, nil
}
