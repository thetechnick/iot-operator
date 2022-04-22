package shelly25rollerclient

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/thetechnick/iot-operator/internal/clients"
)

type Client struct {
	*clients.Client
}

func NewClient(opts clients.ClientOption) *Client {
	return &Client{
		Client: clients.NewClient(opts),
	}
}

func (c *Client) Status(
	ctx context.Context,
) (res Status, err error) {
	return res, c.Do(
		ctx, http.MethodGet, "roller/0", nil, nil, &res)
}

func (c *Client) ToPosition(
	ctx context.Context,
	position int,
) (res Status, err error) {
	return res, c.Do(
		ctx, http.MethodGet, "roller/0", url.Values{
			"go":         []string{"to_pos"},
			"roller_pos": []string{strconv.Itoa(position)},
		}, nil, &res,
	)
}

type Status struct {
	State           State      `json:"state"`
	Power           float64    `json:"power"`
	IsValid         bool       `json:"isValid"`
	SafetySwitch    bool       `json:"safetySwitch"`
	OverTemperature bool       `json:"overtemperature"`
	StopReason      StopReason `json:"stop_reason"`
	LastDirection   Direction  `json:"last_direction"`
	CurrentPos      int        `json:"current_pos"`
	Calibrating     bool       `json:"calibrating"`
	Positioning     bool       `json:"positioning"`
}

type State string

const (
	StateStop  State = "stop"
	StateOpen  State = "open"
	StateClose State = "close"
)

type StopReason string

const (
	StopReasonNormal       StopReason = "normal"
	StopReasonSafetySwitch StopReason = "safety_switch"
	StopReasonObstacle     StopReason = "obstacle"
	StopReasonOverpower    StopReason = "overpower"
)

type Direction string

const (
	DirectionOpen  Direction = "open"
	DirectionClose Direction = "close"
)
