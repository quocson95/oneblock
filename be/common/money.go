package common

type Money struct {
	// Status        string           `json:"status"`
	// ChartSeries   []ChartSery      `json:"chart_series"`
	// Chart         Chart            `json:"chart"`
	// RecessionData []RecessionDatum `json:"recession_data"`
	// UserNotices   []interface{}    `json:"user_notices"`
	Observations [][][]float64 `json:"observations"`
}

type Chart struct {
	Labels              Labels      `json:"labels"`
	Cosd                string      `json:"cosd"`
	Coed                string      `json:"coed"`
	MinDate             string      `json:"min_date"`
	MaxDate             string      `json:"max_date"`
	Frequency           string      `json:"frequency"`
	Width               string      `json:"width"`
	Height              int64       `json:"height"`
	Drp                 int64       `json:"drp"`
	Stacking            interface{} `json:"stacking"`
	Txtcolor            string      `json:"txtcolor"`
	Mode                string      `json:"mode"`
	Ts                  string      `json:"ts"`
	TTS                 string      `json:"tts"`
	Fo                  string      `json:"fo"`
	XScale              string      `json:"x_scale"`
	Trc                 int64       `json:"trc"`
	NT                  int64       `json:"nt"`
	Thu                 int64       `json:"thu"`
	Bgcolor             string      `json:"bgcolor"`
	GraphBgcolor        string      `json:"graph_bgcolor"`
	ShowLegend          string      `json:"showLegend"`
	ShowAxisTitles      string      `json:"showAxisTitles"`
	ZoomType            string      `json:"zoomType"`
	ShowTooltip         string      `json:"showTooltip"`
	ChartType           string      `json:"chartType"`
	RecessionBars       string      `json:"recession_bars"`
	ShowNavigator       string      `json:"showNavigator"`
	AvailableChartTypes []string    `json:"available_chart_types"`
	LogScales           LogScales   `json:"log_scales"`
	AvailableStacking   []string    `json:"available_stacking"`
	LegacyURL           string      `json:"legacy_url"`
	Piedate             interface{} `json:"piedate"`
	LastModified        string      `json:"lastModified"`
	ObsFetch            bool        `json:"obsFetch"`
	DataMin             int64       `json:"data_min"`
	DataMax             int64       `json:"data_max"`
	DataNumObs          int64       `json:"data_num_obs"`
}

type Labels struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	LeftAxis       string `json:"left_axis"`
	RightAxis      string `json:"right_axis"`
	BottomAxis     string `json:"bottom_axis"`
	BubbleSizeAxis string `json:"bubble-size_axis"`
	Footer         string `json:"footer"`
}

type LogScales struct {
	Left       bool `json:"left"`
	Right      bool `json:"right"`
	Bottom     bool `json:"bottom"`
	BubbleSize bool `json:"bubble-size"`
}

type ChartSery struct {
	Type                               string                                    `json:"type"`
	LineIndex                          int64                                     `json:"line_index"`
	LegendIndex                        int64                                     `json:"legendIndex"`
	Title                              string                                    `json:"title"`
	AvailableFormulaTransformations    map[string]AvailableFormulaTransformation `json:"available_formula_transformations"`
	LineColor                          string                                    `json:"line_color"`
	LineStyle                          string                                    `json:"line_style"`
	Lw                                 int64                                     `json:"lw"`
	MarkType                           string                                    `json:"mark_type"`
	HideMarks                          bool                                      `json:"hide_marks"`
	Mw                                 int64                                     `json:"mw"`
	Scale                              string                                    `json:"scale"`
	DecimalPlaces                      string                                    `json:"decimal_places"`
	Frequency                          string                                    `json:"frequency"`
	Fq                                 string                                    `json:"fq"`
	AvailableColors                    map[string]string                         `json:"available_colors"`
	AvailableFams                      AvailableFams                             `json:"available_fams"`
	Fam                                string                                    `json:"fam"`
	AvailableFqs                       []string                                  `json:"available_fqs"`
	Fml                                string                                    `json:"fml"`
	Fgst                               string                                    `json:"fgst"`
	Fgsnd                              string                                    `json:"fgsnd"`
	AllFredSeriesHaveSameFrequency     bool                                      `json:"all_fred_series_have_same_frequency"`
	HasFredSeriesWithNbdTransformation bool                                      `json:"has_fred_series_with_nbd_transformation"`
	Cosd                               string                                    `json:"cosd"`
	Coed                               string                                    `json:"coed"`
	MinDate                            string                                    `json:"min_date"`
	MaxDate                            string                                    `json:"max_date"`
	YearRange                          int64                                     `json:"year_range"`
	Ost                                int64                                     `json:"ost"`
	Oet                                int64                                     `json:"oet"`
	AvailableMmas                      []int64                                   `json:"available_mmas"`
	Mma                                int64                                     `json:"mma"`
	GraphSeriesIDS                     []string                                  `json:"graph_series_ids"`
	SeriesObjects                      SeriesObjects                             `json:"series_objects"`
	Lsv                                interface{}                               `json:"lsv"`
	Lev                                interface{}                               `json:"lev"`
	ObservationGroupingApproximation   string                                    `json:"observation_grouping_approximation"`
	ChartKey                           string                                    `json:"chart_key"`
}

type AvailableFams struct {
	Average     string `json:"Average"`
	Sum         string `json:"Sum"`
	EndOfPeriod string `json:"End of Period"`
}

type AvailableFormulaTransformation struct {
	Full  string `json:"full"`
	Short string `json:"short"`
}

type SeriesObjects struct {
	A A `json:"a"`
}

type A struct {
	SeriesID                         string                           `json:"series_id"`
	Title                            string                           `json:"title"`
	Season                           string                           `json:"season"`
	SeasonShort                      string                           `json:"season_short"`
	Frequency                        string                           `json:"frequency"`
	FrequencyShort                   string                           `json:"frequency_short"`
	UnitsShort                       string                           `json:"units_short"`
	Keywords                         string                           `json:"keywords"`
	Notes                            string                           `json:"notes"`
	AllObsTransformations            AbbreviatedAllObsTransformations `json:"all_obs_transformations"`
	AbbreviatedAllObsTransformations AbbreviatedAllObsTransformations `json:"abbreviated_all_obs_transformations"`
	ObsTransformations               AbbreviatedAllObsTransformations `json:"obs_transformations"`
	ValidStartDate                   string                           `json:"valid_start_date"`
	ValidEndDate                     string                           `json:"valid_end_date"`
	VintageDate                      string                           `json:"vintage_date"`
	AvailableRevisionDates           []string                         `json:"available_revision_dates"`
	RevisionDate                     string                           `json:"revision_date"`
	RelativeVintage                  interface{}                      `json:"relative_vintage"`
	Nd                               string                           `json:"nd"`
	StepLine                         string                           `json:"step_line"`
	Transformation                   string                           `json:"transformation"`
	AvailableUnits                   AbbreviatedAllObsTransformations `json:"available_units"`
	Units                            string                           `json:"units"`
	MinValidStartDate                string                           `json:"min_valid_start_date"`
	MaxValidStartDate                interface{}                      `json:"max_valid_start_date"`
	MinObsStartDate                  string                           `json:"min_obs_start_date"`
	MaxObsStartDate                  string                           `json:"max_obs_start_date"`
	LastUpdated                      string                           `json:"last_updated"`
}

type AbbreviatedAllObsTransformations struct {
	Lin string  `json:"lin"`
	Cap *string `json:"cap,omitempty"`
	Chg string  `json:"chg"`
	Ch1 string  `json:"ch1"`
	Pch string  `json:"pch"`
	Pc1 string  `json:"pc1"`
	Pca string  `json:"pca"`
	Cch string  `json:"cch"`
	Cca string  `json:"cca"`
	Log *string `json:"log,omitempty"`
}

type RecessionDatum struct {
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
