package domain

// ProductInfo represents extracted product information from e-commerce sites
type ProductInfo struct {
	Name     string `json:"name"`
	Brand    string `json:"brand,omitempty"`
	Size     string `json:"size,omitempty"`
	UPC      string `json:"upc,omitempty"`
	Category string `json:"category,omitempty"`
	Retailer string `json:"retailer"` // e.g., "walmart"
	URL      string `json:"url"`
}

// MatchResult represents the result of a product matching operation
type MatchResult struct {
	FdcID         string  `json:"fdcId"`
	Description   string  `json:"description"`
	MatchScore    float64 `json:"matchScore"`
	MatchedTokens []string `json:"matchedTokens,omitempty"`
}
