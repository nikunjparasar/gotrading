package main

import (
	"encoding/json"
	"exchange/orderbook"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler
	ex := NewExchange()

	e.GET("book/:ticker", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	e.Start(":3000")
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
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

type CancelOrderRequest struct {
	Bid bool
	ID  int64
}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[TICKER_ETH]
	order := ob.Orders()[int64(id)]
	ob.CancelOrder(order)

	return c.JSON(http.StatusOK, map[string]any{"msg": "order cancelled"})
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
		vol := order.Size()
		matches := ob.PlaceMarketOrder(order)

		averagePrice := 0.0
		for i := 0; i < len(matches); i++ {
			averagePrice += (matches[i].Price() * matches[i].Size())
		}
		averagePrice /= float64(vol)
		message := fmt.Sprintf("MARKET ORDER PLACED, average price: %.v", averagePrice)

		return c.JSON(http.StatusOK, map[string]any{"msg": message})
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
