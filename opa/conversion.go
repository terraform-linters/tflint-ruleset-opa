package opa

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/types"
)

var rangeTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("filename", types.S),
		types.NewStaticProperty("start", posTy),
		types.NewStaticProperty("end", posTy),
	},
	nil,
)

var posTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("line", types.N),
		types.NewStaticProperty("column", types.N),
		types.NewStaticProperty("bytes", types.N),
	},
	nil,
)

func rangeToJSON(rng hcl.Range) map[string]any {
	return map[string]any{
		"filename": rng.Filename,
		"start":    posToJSON(rng.Start),
		"end":      posToJSON(rng.End),
	}
}

func jsonToRange(in any, path string) (hcl.Range, error) {
	rng, err := jsonToObject(in, path)
	if err != nil {
		return hcl.Range{}, err
	}

	filename, err := jsonToString(rng["filename"], fmt.Sprintf("%s.filename", path))
	if err != nil {
		return hcl.Range{}, err
	}
	start, err := jsonToPos(rng["start"], fmt.Sprintf("%s.start", path))
	if err != nil {
		return hcl.Range{}, err
	}
	end, err := jsonToPos(rng["end"], fmt.Sprintf("%s.end", path))
	if err != nil {
		return hcl.Range{}, err
	}

	return hcl.Range{Filename: filename, Start: start, End: end}, nil
}

func posToJSON(pos hcl.Pos) map[string]int {
	return map[string]int{
		"line":   pos.Line,
		"column": pos.Column,
		"bytes":  pos.Byte,
	}
}

func jsonToPos(in any, path string) (hcl.Pos, error) {
	pos, err := jsonToObject(in, path)
	if err != nil {
		return hcl.Pos{}, err
	}

	line, err := jsonToInt(pos["line"], fmt.Sprintf("%s.line", path))
	if err != nil {
		return hcl.Pos{}, err
	}
	column, err := jsonToInt(pos["column"], fmt.Sprintf("%s.column", path))
	if err != nil {
		return hcl.Pos{}, err
	}
	bytes, err := jsonToInt(pos["bytes"], fmt.Sprintf("%s.bytes", path))
	if err != nil {
		return hcl.Pos{}, err
	}

	return hcl.Pos{Line: line, Column: column, Byte: bytes}, nil
}

func jsonToObject(in any, path string) (map[string]any, error) {
	out, ok := in.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s is not object, got %T", path, in)
	}
	return out, nil
}

func jsonToString(in any, path string) (string, error) {
	out, ok := in.(string)
	if !ok {
		return "", fmt.Errorf("%s is not string, got %T", path, in)
	}
	return out, nil
}

func jsonToInt(in any, path string) (int, error) {
	jn, ok := in.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s is not a number, got %T", path, in)
	}
	num, err := jn.Int64()
	if err != nil {
		return 0, err
	}
	return int(num), nil
}
