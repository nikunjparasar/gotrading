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

	e.Start(":3000")
}

// define the market
type Ticker string
type OrderType string

const (
	TICKER_ETH Ticker = "ETH"

	LIMIT_ORDER  OrderType = "LIMIT"
	MARKET_ORDER OrderType = "MARKET"

	SELL bool = false
	BUY  bool = true
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
	Buy    bool
	Size   float64
	Price  float64
	Ticker Ticker
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	ticker := Ticker(placeOrderData.Ticker)
	ob := ex.orderbooks[ticker]
	order := orderbook.NewOrder(placeOrderData.Buy, placeOrderData.Size)
	ob.PlaceLimitOrder(placeOrderData.Price, order)

	return c.JSON(200, map[string]any{"msg": "order placed"})
}

// JSON representation
type Order struct {
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderbookData struct {
	Asks []*Order
	Bids []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	ticker := Ticker(c.Param("ticker"))
	ob, ok := ex.orderbooks[ticker]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "ticker not found"})
	}

	orderbookData := OrderbookData{
		Asks: []*Order{},
		Bids: []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders() {
			o := Order{
				Price:     limit.Price(),
				Size:      order.Size(),
				Bid:       order.Buy(),
				Timestamp: order.Timestamp(),
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}
