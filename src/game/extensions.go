package game

type Settings struct {
	FuturesMode  bool `json:"futures,omitempty"`
	SplurgesMode bool `json:"splurges,omitempty"`
}

func (s *Settings) String() (str string) {
	if s.FuturesMode {
		str += " Futures"
	}
	if s.SplurgesMode {
		str += " Splurges"
	}
	return str[1:]
}

type Future struct {
	Src int `json:"source"`
	Dst int `json:"target"`
}
