package gps

// Address is the address view of a geographical location
type Address struct {
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"ciso,omitempty"`
	City        string `json:"city,omitempty"`
	Zip         string `json:"zip,omitempty"`
	County      string `json:"county,omitempty"`
}
