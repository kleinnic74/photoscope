package gps

// Address is the address view of a geographical location
type Address struct {
	Country
	Place
	County string `json:"county,omitempty"`
}

type Country struct {
	Country string `json:"country,omitempty"`
	Code    string `json:"ciso,omitempty"`
}

type Place struct {
	City string `json:"city,omitempty"`
	Zip  string `json:"zip,omitempty"`
}
