package common

type CACertificateParams struct {
	CommonName    string `json:"commonName,omitzero"`
	CountryCode   string `json:"countryCode,omitzero"`
	EmailAddress  string `json:"emailAddress,omitzero"`
	Locality      string `json:"locality,omitzero"`
	Organization  string `json:"organization,omitzero"`
	Unit          string `json:"organizationUnit,omitzero"`
	State         string `json:"state,omitzero"`
	Province      string `json:"province,omitzero"`
	StreetAddress string `json:"streetAddress,omitzero"`
	PostalCode    string `json:"postalCode,omitzero"`
	IssueDate     string `json:"issueDate,omitzero"`
	ExpiryDate    string `json:"expiryDate,omitzero"`
}

type PrivateCAConfig struct {
	CACerificateParams CACertificateParams `json:"privateCA"`
}

type TLSConfig struct {
	CAConfig any `json:"tls"`
}

type BaseInfraConfig struct {
	RootDomain string          `json:"rootDomain"`
	TLS        PrivateCAConfig `json:"tls"`
}
