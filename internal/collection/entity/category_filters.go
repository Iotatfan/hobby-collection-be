package entity

// CategoryFilterResponse describes the complete, unfiltered shelf population.
type CategoryFilterResponse struct {
	Category      string                      `json:"category"`
	TotalCount    int64                       `json:"total_count"`
	Grades        []GunplaGradeFilterResponse `json:"grades"`
	Scales        []ScaleResponse             `json:"scales"`
	ReleaseTypes  []ReleaseTypeFilterResponse `json:"release_types"`
	Manufacturers []ManufacturerResponse      `json:"manufacturers"`
	Series        []SeriesResponse            `json:"series"`
}
