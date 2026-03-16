package entity

import "strings"

type Vinculo struct {
	ID      string `json:"id" firestore:"id,omitempty"`
	EmailA  string `json:"email_a" firestore:"email_a"`
	EmailB  string `json:"email_b" firestore:"email_b"`
}

func (v Vinculo) GetOutroEmail(meuEmail string) string {
	if strings.ToLower(v.EmailA) == strings.ToLower(meuEmail) {
		return v.EmailB
	}
	return v.EmailA
}
