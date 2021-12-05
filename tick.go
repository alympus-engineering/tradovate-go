package tradovate

import (
	"encoding/json"
	"time"
)

type ChartRequest struct {
	Symbol 				string `json:"symbol"`
	ChartDescription 	ChartDescription `json:"chartDescription"`
	TimeRange 			TimeRange `json:"timeRange"`
}

type ChartDescription struct {
	// Tick, DailyBar, MinuteBar, Custom, DOM
	UnderlyingType 	string 	`json:"underlyingType,omitempty"`
	ElementType    	int 	`json:"elementType,omitempty"`
	// Volume, Range, UnderlyingUnits, Renko, MomentumRange, PointAndFigure, OFARange
	ElementSizeUnit string 	`json:"elementSizeUnit,omitempty"`
	WithHistogram 	bool 	`json:"withHistogram,omitempty"`
}

type TimeRange struct {
	ClosestTimestamp 	time.Time `json:"closestTimestamp,omitempty"`
	ClosestTickId		int `json:"closestTickId,omitempty"`
	AsFarAsTimestamp 	time.Time `json:"asFarAsTimestamp,omitempty"`
	AsMuchAsElement		int `json:"asMuchAsElement,omitempty"`
}

func (c *Client) GetHistoricalTickData(symbol string, start time.Time, end time.Time) error {
	chartReq := ChartRequest {
		Symbol: symbol,
		ChartDescription: ChartDescription{
			UnderlyingType:  "Tick",
			ElementType:     1,
			ElementSizeUnit: "",
			WithHistogram:   false,
		},
		TimeRange: TimeRange{
			ClosestTimestamp: end,
			AsFarAsTimestamp: start,
		},
	}

	b, err := json.Marshal(chartReq)

	msg, err := c.Send(GET_CHART, "", string(b), 10)

	if msg.Status != 200 {

	}

	return err
}