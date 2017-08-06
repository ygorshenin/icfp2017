package common

import "fmt"

type ClaimMove struct {
	Punter int `json:"punter"`
	Source int `json:"source"`
	Target int `json:"target"`
}

func (m *ClaimMove) String() string {
	return fmt.Sprintf("Punter=%v, Claim River=(%v, %v)", m.Punter, m.Source, m.Target)
}

type PassMove struct {
	Punter int `json:"punter"`
}

func (m *PassMove) String() string {
	return fmt.Sprintf("Punter=%v, Pass", m.Punter)
}

type SplurgeMove struct {
	Punter int   `json:"punter"`
	Route  []int `json:"route,omitempty"`
}

func (m *SplurgeMove) String() string {
	return fmt.Sprintf("Punter=%v, Splurge Route=%v", m.Punter, m.Route)
}

type Move struct {
	Claim   *ClaimMove   `json:"claim,omitempty"`
	Pass    *PassMove    `json:"pass,omitempty"`
	Splurge *SplurgeMove `json:"splurge,omitempty"`
	State   *PlayerProxy `json:"state"`
}

func (m *Move) String() string {
	if m.Claim != nil {
		return m.Claim.String()
	}
	if m.Pass != nil {
		return m.Pass.String()
	}
	if m.Splurge != nil {
		return m.Splurge.String()
	}
	return "Bad Move"
}
