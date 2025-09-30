package model

type Pos struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  int     `json:"z"`
	A  int     `json:"a"`
	S  int     `json:"s"`
	St int     `json:"st"`
}

type Data struct {
	ID      int64                  `json:"id"`
	T       int                    `json:"t"`
	ST      int                    `json:"st"`
	Pos     Pos                    `json:"pos"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Address string                 `json:"address,omitempty"`
}
