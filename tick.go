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
	ElementSize		int 	`json:"elementSize,omitempty"`
	// Volume, Range, UnderlyingUnits, Renko, MomentumRange, PointAndFigure, OFARange
	ElementSizeUnit string 	`json:"elementSizeUnit,omitempty"`
	WithHistogram 	bool 	`json:"withHistogram,omitempty"`
}

type TimeRange struct {
	ClosestTimestamp 	Time `json:"closestTimestamp,omitempty"`
	ClosestTickId		int `json:"closestTickId,omitempty"`
	AsFarAsTimestamp 	Time `json:"asFarAsTimestamp,omitempty"`
	AsMuchAsElement		int `json:"asMuchAsElement,omitempty"`
}

type Subscription struct {
	Mode 			string `json:"mode"`
	HistoricalId 	int `json:"historicalId"`
	RealtimeId 		int `json:"realtimeId"`
}

type Chart struct {

}

type Tick struct {
	SubscriptionId 	string 		`json:"id"`
	Source			string 		`json:"source"`
	Time			time.Time 	`json:"time"`
	Price			float64 		`json:"price"`
	Bid 			float64 		`json:"bid"`
	Ask				float64 		`json:"ask"`
	AskSize			int	   		`json:"ask_size"`
	BidSize			int			`json:"bid_size"`
}

func (c *Client) GetHistoricalTickData(symbol string, start time.Time, end time.Time) (Subscription, error) {
	chartReq := ChartRequest {
		Symbol: symbol,
		ChartDescription: ChartDescription{
			UnderlyingType:  "Tick",
			ElementSize: 	 1,
			ElementSizeUnit: "UnderlyingUnits",
			WithHistogram:   false,
		},
		TimeRange: TimeRange{
			ClosestTimestamp: Time{end},
			AsFarAsTimestamp: Time{start},
		},
	}

	b, err := json.Marshal(chartReq)

	msg, err := c.Send(GET_CHART, "", string(b), 10)

	if msg.Status != 200 {

	}

	var subscription Subscription
	err = json.Unmarshal(msg.Data, &subscription)

	return subscription, err
}

func (c *Client) CancelHistoricalTickData(id int) error {
	cancelChartReq := struct {
		SubscriptionId int `json:"subscriptionId"`
	} {
		SubscriptionId: id,
	}

	b, err := json.Marshal(cancelChartReq)

	if err != nil {

	}

	msg, err := c.Send(CANCEL_CHART, "", string(b), 10)

	if msg.Status != 200 {

	}

	return err
}

func (c *Client) UnsubscribeRealtimeData(id int) error {
	cancelChartReq := struct {
		SubscriptionId int `json:"subscriptionId"`
	} {
		SubscriptionId: id,
	}

	b, err := json.Marshal(cancelChartReq)

	if err != nil {

	}

	msg, err := c.Send(CANCEL_CHART, "", string(b), 10)

	if msg.Status != 200 {

	}

	return err
}