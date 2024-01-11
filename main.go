package main

import (
	"encoding/json"
	"exchange/orderbook"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	ex := NewExchange()

	e.GET("book/:ticker", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	e.Start(":3000")
}

// define the market
type Ticker string
type OrderType string

const (
	TICKER_ETH Ticker = "ETH"

	LIMIT_ORDER  OrderType = "LIMIT"
	MARKET_ORDER OrderType = "MARKET"

	BUY  = "BUY"
	SELL = "SELL"
)

type Exchange struct {
	orderbooks map[Ticker]*orderbook.Orderbook
}

func NewExchange() *Exchange {

	orderbooks := make(map[Ticker]*orderbook.Orderbook)
	orderbooks[TICKER_ETH] = orderbook.NewOrderbook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type PlaceOrderRequest struct {
	// public fields
	Type   OrderType //limit or market
	Action string
	Size   float64
	Price  float64
	Ticker Ticker
}

// type CancelOrderRequest struct {
// 	Bid bool
// 	ID  int64
// }

func (ex *Exchange) cancelOrder(c echo.Context) error {
	// id := c.Param("id")

	// if err := json.NewDecoder(c.Request().Body).Decode(&id); err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid id"})
	// }

	return nil
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	ticker := Ticker(placeOrderData.Ticker)
	ob := ex.orderbooks[ticker]

	order := orderbook.NewOrder(placeOrderData.Action, placeOrderData.Size)

	if placeOrderData.Type == LIMIT_ORDER {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
		return c.JSON(http.StatusOK, map[string]any{"msg": "LIMIT ORDER PLACED"})
	}

	if placeOrderData.Type == MARKET_ORDER {
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(http.StatusOK, map[string]any{"matches": len(matches)})
	}

	return nil
}

// JSON representation
type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Action    string
	Timestamp int64
}

type OrderbookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*Order
	Bids           []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	ticker := Ticker(c.Param("ticker"))
	ob, ok := ex.orderbooks[ticker]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "ticker not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolume: ob.BidsTotalVolume(),
		TotalAskVolume: ob.AsksTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks().Limits {
		for _, order := range limit.Orders() {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price(),
				Size:      order.Size(),
				Action:    order.Action(),
				Timestamp: order.Timestamp(),
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}
	for _, limit := range ob.Bids().Limits {
		for _, order := range limit.Orders() {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price(),
				Size:      order.Size(),
				Action:    order.Action(),
				Timestamp: order.Timestamp(),
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}
